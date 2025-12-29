package services

import (
	"context"
	"fmt"
	"personal_schedule_service/internal/collection"
	labels_constant "personal_schedule_service/internal/constant/labels"
	notifications_constant "personal_schedule_service/internal/constant/notifications"
	"personal_schedule_service/internal/grpc/mapper"
	"personal_schedule_service/internal/grpc/utils"
	"personal_schedule_service/internal/grpc/validation"
	"personal_schedule_service/internal/repos"
	app_error "personal_schedule_service/pkg/settings/error"
	"personal_schedule_service/proto/common"
	"personal_schedule_service/proto/personal_schedule"
	"time"

	"github.com/thanvuc/go-core-lib/eventbus"
	"github.com/thanvuc/go-core-lib/log"
	"github.com/thanvuc/go-core-lib/mongolib"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type workService struct {
	logger            log.Logger
	workRepo          repos.WorkRepo
	workMapper        mapper.WorkMapper
	mongoConnector    *mongolib.MongoConnector
	validator         validation.WorkValidator
	eventbusConnector *eventbus.RabbitMQConnector
}

type recovertTimes struct {
	TargetStart time.Time
	TargetEnd   time.Time
	SourceStart time.Time
	SourceEnd   time.Time
	TimeShift   time.Duration
}

func (s *workService) UpsertWork(ctx context.Context, req *personal_schedule.UpsertWorkRequest) (*personal_schedule.UpsertWorkResponse, error) {
	requestId := utils.GetRequestIDFromOutgoingContext(ctx)
	if err := s.validator.ValidateUpsertWork(ctx, req); err != nil {
		s.logger.Error("UpsertWork validation failed", requestId, zap.Error(err))
		if ve, ok := err.(*validation.ValidationError); ok {
			return &personal_schedule.UpsertWorkResponse{
				IsSuccess: false,
				Error:     utils.CustomError(ctx, ve.Category, ve.Code, err),
			}, nil
		}
	}

	work, subTasksDB, err := s.workMapper.MapUpsertProtoToModels(req)
	if err != nil {
		s.logger.Error("Failed to map UpsertWorkRequest to models", requestId, zap.Error(err))
		return &personal_schedule.UpsertWorkResponse{
			IsSuccess: false,
			Error:     utils.InternalServerError(ctx, err),
		}, nil
	}

	now := time.Now().UTC()
	if req.Id == nil || *req.Id == "" {
		work.UserID = req.UserId
		work.CreatedAt = now
		work.LastModifiedAt = now

		newID, err := s.workRepo.CreateWork(ctx, work)
		if err != nil {
			s.logger.Error("Failed to create work", requestId, zap.Error(err))
			return &personal_schedule.UpsertWorkResponse{
				IsSuccess: false, Message: "Failed to create work", Error: utils.DatabaseError(ctx, err),
			}, err
		}
		work.ID = newID

	} else {
		workID, _ := bson.ObjectIDFromHex(*req.Id)

		if err := s.workRepo.UpdateWork(ctx, workID, work); err != nil {
			s.logger.Error("Failed to update work", requestId, zap.Error(err))
			return &personal_schedule.UpsertWorkResponse{
				IsSuccess: false, Message: "Failed to update work", Error: utils.DatabaseError(ctx, err),
			}, err
		}
		work.ID = workID
	}
	if err := s.syncSubTasks(ctx, work.ID, subTasksDB); err != nil {
		s.logger.Error("Failed to sync sub-tasks", requestId, zap.Error(err))
		return &personal_schedule.UpsertWorkResponse{
			IsSuccess: false, Message: "Failed to sync sub-tasks (Work was upserted but tasks failed)", Error: utils.DatabaseError(ctx, err),
		}, err
	}

	if len(req.Notifications) > 0 {
		if err := s.sendNotificationEvent(ctx, req, work.ID.Hex()); err != nil {
			s.logger.Error("Failed to send notification event", requestId, zap.Error(err))
			return &personal_schedule.UpsertWorkResponse{
				IsSuccess: false,
				Message:   "Failed to send notification event (Work was upserted but notification failed)",
				Error:     utils.CustomError(ctx, common.ErrorCode_ERROR_CODE_INTERNAL_ERROR, app_error.NotificationCannotBeSent, err),
			}, err
		}
	}

	return &personal_schedule.UpsertWorkResponse{
		IsSuccess: true,
		Message:   "Work upserted successfully",
	}, nil
}

func (s *workService) syncSubTasks(ctx context.Context, workID bson.ObjectID, payloadTasks []collection.SubTask) error {
	existingTasks, err := s.workRepo.GetSubTasksByWorkID(ctx, workID)
	if err != nil {
		return err
	}

	existingTaskMap := make(map[bson.ObjectID]bool)
	for _, task := range existingTasks {
		existingTaskMap[task.ID] = true
	}

	var operations []mongo.WriteModel
	now := time.Now()

	for _, task := range payloadTasks {
		task.WorkID = workID
		if task.ID.IsZero() {
			task.ID = bson.NewObjectID()
			task.CreatedAt = now
			task.LastModifiedAt = now
			operations = append(operations, mongo.NewInsertOneModel().SetDocument(task))
		} else {
			delete(existingTaskMap, task.ID)
			operations = append(operations, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": task.ID, "work_id": workID}).
				SetUpdate(bson.M{"$set": bson.M{
					"name":             task.Name,
					"is_completed":     task.IsCompleted,
					"last_modified_at": now,
				}}))
		}
	}

	for taskID := range existingTaskMap {
		operations = append(operations, mongo.NewDeleteOneModel().SetFilter(bson.M{"_id": taskID, "work_id": workID}))
	}

	if len(operations) > 0 {
		_, err = s.workRepo.BulkWriteSubTasks(ctx, operations)
		return err
	}

	return nil
}

func (s *workService) sendNotificationEvent(ctx context.Context, req *personal_schedule.UpsertWorkRequest, workId string) error {
	// Prepare event data
	notifications := common.Notifications{}
	for _, notification := range req.Notifications {
		id := utils.Ternary(notification.Id != nil, *notification.Id, "")
		message := utils.Ternary(req.ShortDescriptions != nil, *req.ShortDescriptions, utils.Ternary(req.DetailedDescription != nil, *req.DetailedDescription, req.Name))
		link := utils.Ternary(notification.Link != nil, *notification.Link+workId, "")
		notificationPayload := &common.Notification{
			Id:              &id,
			Title:           req.Name,
			Message:         message,
			SenderId:        req.UserId,
			ReceiverIds:     []string{req.UserId},
			IsRead:          false,
			Link:            &link,
			IsActive:        notification.IsActive,
			TriggerAt:       &notification.TriggerAt,
			IsSendMail:      notification.IsSendMail,
			CorrelationId:   workId,
			CorrelationType: common.NOTIFICATION_TYPE_SCHEDULED_NOTIFICATION,
			ImageUrl:        notification.ImgUrl,
		}
		notifications.Notifications = append(notifications.Notifications, notificationPayload)
	}
	notificationsPayload, err := proto.Marshal(&notifications)
	if err != nil {
		s.logger.Error("Failed to marshal notifications payload", "", zap.Error(err))
		return err
	}

	// Publish event to RabbitMQ
	if err := s.publishNotifications(ctx, notificationsPayload); err != nil {
		s.logger.Error("Failed to publish notifications event", "", zap.Error(err))
		return err
	}

	return nil
}

func (s *workService) publishNotifications(ctx context.Context, payload []byte) error {
	publisher := eventbus.NewPublisher(
		s.eventbusConnector,
		notifications_constant.NOTIFICATION_EXCHANGE,
		eventbus.ExchangeTypeTopic,
		nil,
		nil,
		false,
	)

	requestId := utils.GetRequestIDFromOutgoingContext(ctx)

	err := publisher.Publish(ctx, requestId, []string{notifications_constant.NOTIFICATION_ROUTING_KEY}, payload, nil)
	if err != nil {
		s.logger.Error("Failed to publish notification event", "", zap.Error(err))
		return err
	}

	return nil
}

func (s *workService) GetWorks(ctx context.Context, req *personal_schedule.GetWorksRequest) (*personal_schedule.GetWorksResponse, error) {
	aggWorks, totalWorks, err := s.workRepo.GetWorks(ctx, req)
	if err != nil {
		s.logger.Error("Failed to get works", "", zap.Error(err))
		return &personal_schedule.GetWorksResponse{
			Works:      []*personal_schedule.Work{},
			TotalWorks: 0,
			Error:      utils.InternalServerError(ctx, err),
		}, nil
	}

	overdueLabel, err := s.workRepo.GetLabelByKey(ctx, labels_constant.LabelOverDue)
	if err != nil {
		s.logger.Error("Failed to get overdue label", "", zap.Error(err))
	}

	timeNow := time.Now()

	for i := range aggWorks {
		if aggWorks[i].EndDate.Before(timeNow) {
			aggWorks[i].Overdue = []collection.Label{*overdueLabel}
		}
	}

	protoWorks := s.workMapper.ConvertAggregatedWorksToProto(aggWorks)

	return &personal_schedule.GetWorksResponse{
		Works:      protoWorks,
		TotalWorks: int32(totalWorks),
		Error:      nil,
	}, nil

}

func (s *workService) GetWork(ctx context.Context, req *personal_schedule.GetWorkRequest) (*personal_schedule.GetWorkResponse, error) {
	workID, err := bson.ObjectIDFromHex(req.WorkId)
	if err != nil {
		s.logger.Warn("Invalid work ID format", "", zap.String("work_id", req.WorkId), zap.Error(err))
		return &personal_schedule.GetWorkResponse{
			Work:  nil,
			Error: utils.InternalServerError(ctx, err),
		}, nil
	}

	work, err := s.workRepo.GetAggregatedWorkByID(ctx, workID)
	if err != nil {
		s.logger.Error("Error fetching work from repo", "err", zap.Error(err))
		return &personal_schedule.GetWorkResponse{
			Work:  nil,
			Error: utils.DatabaseError(ctx, err),
		}, err
	}

	if work == nil {
		s.logger.Info("Work not found", "", zap.String("work_id", req.WorkId))
		return &personal_schedule.GetWorkResponse{
			Work:  nil,
			Error: utils.NotFoundError(ctx, err),
		}, nil
	}

	if work.UserID != req.UserId {
		s.logger.Warn("Forbidden: user does not own this work", "", zap.String("work_id", req.WorkId), zap.String("user_id", req.UserId))
		return &personal_schedule.GetWorkResponse{
			Work:  nil,
			Error: utils.PermissionDeniedError(ctx, err),
		}, nil
	}

	subTasksDB, err := s.workRepo.GetSubTasksByWorkID(ctx, workID)
	if err != nil {
		s.logger.Error("Error fetching work sub-tasks from repo", "err", zap.Error(err))
		return &personal_schedule.GetWorkResponse{
			Work:  nil,
			Error: utils.DatabaseError(ctx, err),
		}, err
	}

	protoWork := s.workMapper.MapAggregatedToWorkDetailProto(*work, subTasksDB)

	return &personal_schedule.GetWorkResponse{
		Work:  protoWork,
		Error: nil,
	}, nil
}

func (s *workService) DeleteWork(ctx context.Context, req *personal_schedule.DeleteWorkRequest) (*personal_schedule.DeleteWorkResponse, error) {
	workID, err := bson.ObjectIDFromHex(req.WorkId)
	if err != nil {
		s.logger.Warn("Invalid work ID format", "", zap.String("work_id", req.WorkId), zap.Error(err))
		return &personal_schedule.DeleteWorkResponse{
			Success: false,
			Error:   utils.InternalServerError(ctx, err),
		}, nil
	}

	work, err := s.workRepo.GetWorkByID(ctx, workID)
	if err != nil {
		s.logger.Error("Error fetching work from repo", "err", zap.Error(err))
		return &personal_schedule.DeleteWorkResponse{
			Success: false,
			Error:   utils.DatabaseError(ctx, err),
		}, err
	}
	if work == nil {
		s.logger.Info("Goal not found", "", zap.String("goal_id", req.WorkId))
		return nil, nil
	}

	if work.UserID != req.UserId {
		s.logger.Warn("Forbidden: user does not own this work", "", zap.String("work_id", req.WorkId), zap.String("user_id", req.UserId))
		return &personal_schedule.DeleteWorkResponse{
			Success: false,
			Error:   utils.PermissionDeniedError(ctx, err),
		}, nil
	}

	err = s.workRepo.DeleteSubTaskByWorkID(ctx, workID)
	if err != nil {
		s.logger.Error("Error deleting sub-tasks by work ID", "err", zap.Error(err))
		return &personal_schedule.DeleteWorkResponse{
			Success: false,
			Error:   utils.DatabaseError(ctx, err),
		}, err
	}
	err = s.workRepo.DeleteWork(ctx, workID)
	if err != nil {
		s.logger.Error("Error deleting work", "err", zap.Error(err))
		return &personal_schedule.DeleteWorkResponse{
			Success: false,
			Error:   utils.DatabaseError(ctx, err),
		}, err
	}
	return &personal_schedule.DeleteWorkResponse{
		Success: true,
	}, nil

}

func (s *workService) parseRecoverTimes(req *personal_schedule.GetRecoveryWorksRequest) (*recovertTimes, error) {
	targetStart := time.UnixMilli(req.TargetDate)
	sourceStart := time.UnixMilli(req.SourceDate)

	return &recovertTimes{
		TargetStart: targetStart,
		TargetEnd:   utils.EndOfDay(targetStart),
		SourceStart: sourceStart,
		SourceEnd:   utils.EndOfDay(sourceStart),
		TimeShift:   targetStart.Sub(sourceStart),
	}, nil
}

func (s *workService) cloneSubTasks(oldSubs []collection.SubTask, newWorkID bson.ObjectID, now time.Time) []collection.SubTask {
	var result []collection.SubTask
	for _, s := range oldSubs {
		result = append(result, collection.SubTask{
			ID:             bson.NewObjectID(),
			Name:           s.Name,
			IsCompleted:    false,
			WorkID:         newWorkID,
			CreatedAt:      now,
			LastModifiedAt: now,
		})
	}
	return result
}

func (s *workService) cloneSingleWork(ctx context.Context, oldWork repos.AggregatedWork, times *recovertTimes, draft *collection.Label) (newWork collection.Work, subTasks []collection.SubTask, err error) {
	now := time.Now().UTC()
	newEnd := oldWork.EndDate.Add(times.TimeShift)
	var newStart *time.Time
	if oldWork.StartDate != nil {
		t := oldWork.StartDate.Add(times.TimeShift)
		newStart = &t
	}

	newWorkID := bson.NewObjectID()
	oldSubs, err := s.workRepo.GetSubTasksByWorkID(ctx, oldWork.ID)
	if err != nil {
		s.logger.Error("Failed to get subtasks for work", "", zap.Error(err))
		return newWork, nil, err
	}
	subTasks = s.cloneSubTasks(oldSubs, newWorkID, now)

	inProgress, err := s.workRepo.GetLabelByKey(ctx, labels_constant.LabelInProgress)
	if err != nil {
		s.logger.Error("Failed to get in-progress label", "", zap.Error(err))
		return newWork, nil, err
	}

	var goalID *bson.ObjectID
	if len(oldWork.GoalInfo) > 0 {
		gid := oldWork.GoalInfo[0].ID
		goalID = &gid
	}

	newWork = collection.Work{
		ID:                  newWorkID,
		Name:                oldWork.Name,
		ShortDescriptions:   oldWork.ShortDescriptions,
		DetailedDescription: oldWork.DetailedDescription,
		NameNormalized:      oldWork.NameNormalized,
		StartDate:           newStart,
		EndDate:             newEnd,
		UserID:              oldWork.UserID,
		StatusID:            inProgress.ID,
		DifficultyID:        oldWork.Difficulty[0].ID,
		PriorityID:          oldWork.Priority[0].ID,
		TypeID:              oldWork.Type[0].ID,
		CategoryID:          oldWork.Category[0].ID,
		GoalID:              goalID,
		DraftID:             &draft.ID,
		CreatedAt:           now,
		LastModifiedAt:      now,
	}

	return newWork, subTasks, nil
}

func (s *workService) RecoverWorks(ctx context.Context, req *personal_schedule.GetRecoveryWorksRequest) (*personal_schedule.GetRecoveryWorksResponse, error) {
	times, err := s.parseRecoverTimes(req)
	if err != nil {
		s.logger.Error("Failed to parse recover times", "", zap.Error(err))
		return &personal_schedule.GetRecoveryWorksResponse{
			IsSuccess: false,
			Message:   "Failed to parse recover times",
			Error:     utils.InternalServerError(ctx, err),
		}, nil
	}

	draft, err := s.workRepo.GetLabelByKey(ctx, labels_constant.LabelDraft)
	if err != nil {
		s.logger.Error("Failed to get draft label", "", zap.Error(err))
		return nil, err
	}

	if err := s.workRepo.DeleteDraftsByDate(ctx, req.UserId, times.TargetStart, times.TargetEnd); err != nil {
		s.logger.Error("Failed to delete drafts by date", "", zap.Error(err))
		return &personal_schedule.GetRecoveryWorksResponse{
			IsSuccess: false,
			Message:   "Failed to delete drafts by date",
			Error:     utils.DatabaseError(ctx, err),
		}, nil
	}

	sourceWorks, err := s.workRepo.GetAggregatedWorksByDateRangeMs(ctx, req.UserId, times.SourceStart.UnixMilli(), times.SourceEnd.UnixMilli())
	if err != nil {
		s.logger.Error("Failed to get source works", "", zap.Error(err))
		return &personal_schedule.GetRecoveryWorksResponse{
			IsSuccess: false,
			Message:   "Failed to get source works",
			Error:     utils.DatabaseError(ctx, err),
		}, nil
	}

	var worksToInsert []interface{}
	var subTasksToInsert []interface{}

	for _, oldWork := range sourceWorks {
		newWork, subtasks, err := s.cloneSingleWork(ctx, oldWork, times, draft)
		if err != nil {
			s.logger.Error("Failed to clone single work", "", zap.Error(err))
			return &personal_schedule.GetRecoveryWorksResponse{
				IsSuccess: false,
				Message:   "Failed to clone single work",
				Error:     utils.DatabaseError(ctx, err),
			}, nil
		}

		worksToInsert = append(worksToInsert, newWork)

		for _, st := range subtasks {
			subTasksToInsert = append(subTasksToInsert, st)
		}

	}
	if err := s.workRepo.BulkInsertWorks(ctx, worksToInsert); err != nil {
		return &personal_schedule.GetRecoveryWorksResponse{
			IsSuccess: false,
			Message:   "Failed to insert recovered works",
			Error:     utils.DatabaseError(ctx, err),
		}, nil
	}

	if len(subTasksToInsert) > 0 {
		_ = s.workRepo.BulkInsertSubTasks(ctx, subTasksToInsert)
	}

	return &personal_schedule.GetRecoveryWorksResponse{
		IsSuccess: true,
		Message:   "Works recovered successfully",
		Error:     nil,
	}, nil
}

func (s *workService) UpdateWorkLabel(ctx context.Context, req *personal_schedule.UpdateWorkLabelRequest) (*personal_schedule.UpdateWorkLabelResponse, error) {
	workID, err := bson.ObjectIDFromHex(req.WorkId)
	if err != nil {
		s.logger.Warn("Invalid work ID format", "", zap.String("work_id", req.WorkId), zap.Error(err))
		return &personal_schedule.UpdateWorkLabelResponse{
			Error: utils.InternalServerError(ctx, err),
		}, nil
	}

	labelID, err := bson.ObjectIDFromHex(req.LabelId)
	if err != nil {
		s.logger.Warn("Invalid label ID format", "", zap.String("label_id", req.LabelId), zap.Error(err))
		return &personal_schedule.UpdateWorkLabelResponse{
			Error: utils.InternalServerError(ctx, err),
		}, nil
	}

	work, err := s.workRepo.GetWorkByID(ctx, workID)
	if err != nil {
		s.logger.Error("Error fetching work from repo", "err", zap.Error(err))
		return &personal_schedule.UpdateWorkLabelResponse{
			Error: utils.DatabaseError(ctx, err),
		}, nil
	}
	if work == nil {
		errMsg := "work not found"
		s.logger.Info(errMsg, "", zap.String("work_id", req.WorkId))
		return &personal_schedule.UpdateWorkLabelResponse{
			Error: utils.NotFoundError(ctx, err),
		}, nil
	}
	if work.UserID != req.UserId {
		errMsg := "forbidden: user does not own this work"
		s.logger.Warn(errMsg, "", zap.String("work_id", req.WorkId), zap.String("user_id", req.UserId))
		return &personal_schedule.UpdateWorkLabelResponse{
			Error: utils.PermissionDeniedError(ctx, err),
		}, nil
	}

	var fieldName string
	switch req.LabelType {
	case 1:
		fieldName = "type_id"
	case 2:
		fieldName = "status_id"
	case 3:
		fieldName = "difficulty_id"
	case 4:
		fieldName = "priority_id"
	case 5:
		fieldName = "category_id"
	default:
		errMsg := "invalid label type"
		s.logger.Warn(errMsg, "", zap.Int32("label_type", req.LabelType))
		return &personal_schedule.UpdateWorkLabelResponse{
			Error: utils.InternalServerError(ctx, err),
		}, nil
	}
	fmt.Println("Updating work", workID.Hex(), "field", fieldName, "to label", labelID.Hex())

	if err := s.workRepo.UpdateWorkField(ctx, workID, fieldName, labelID); err != nil {
		s.logger.Error("Failed to update work label", "", zap.Error(err))
		return &personal_schedule.UpdateWorkLabelResponse{
			Error: utils.DatabaseError(ctx, err),
		}, nil
	}

	return &personal_schedule.UpdateWorkLabelResponse{
		IsSuccess: true,
		Message:   "Label updated successfully",
	}, nil
}

func (s *workService) CommitRecoveryDrafts(ctx context.Context, req *personal_schedule.CommitRecoveryDraftsRequest) (*personal_schedule.CommitRecoveryDraftsResponse, error) {
	draftID, err := s.workRepo.GetLabelByKey(ctx, labels_constant.LabelDraft)
	if err != nil {
		s.logger.Error("Failed to get draft label", "", zap.Error(err))
		return &personal_schedule.CommitRecoveryDraftsResponse{
			IsSuccess: false,
			Message:   "Failed to get draft label",
			Error:     utils.DatabaseError(ctx, err),
		}, nil
	}

	if len(req.WorkIds) == 0 {
		return &personal_schedule.CommitRecoveryDraftsResponse{
			IsSuccess: true,
		}, nil
	}
	var workObjectIDs []bson.ObjectID
	for _, id := range req.WorkIds {
		oid, err := bson.ObjectIDFromHex(id)
		if err != nil {
			return &personal_schedule.CommitRecoveryDraftsResponse{
				IsSuccess: false,
			}, nil
		}
		workObjectIDs = append(workObjectIDs, oid)
	}

	err = s.workRepo.CommitRecoveryDrafts(ctx, req.UserId, workObjectIDs, draftID.ID)
	if err != nil {
		s.logger.Error("Failed to commit recovery drafts", "", zap.Error(err))
		return &personal_schedule.CommitRecoveryDraftsResponse{
			IsSuccess: false,
			Message:   "Failed to commit recovery drafts",
			Error:     utils.DatabaseError(ctx, err),
		}, nil
	}
	return &personal_schedule.CommitRecoveryDraftsResponse{
		IsSuccess: true,
		Message:   "Recovery drafts committed successfully",
	}, nil
}
