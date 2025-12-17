package services

import (
	"context"
	"fmt"
	"personal_schedule_service/internal/collection"
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

	now := time.Now()
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
	aggWorks, err := s.workRepo.GetWorks(ctx, req)
	if err != nil {
		s.logger.Error("Failed to get works", "", zap.Error(err))
		return &personal_schedule.GetWorksResponse{
			Works: []*personal_schedule.Work{},
			Error: utils.InternalServerError(ctx, err),
		}, nil
	}
	protoWorks := s.workMapper.ConvertAggregatedWorksToProto(aggWorks)

	return &personal_schedule.GetWorksResponse{
		Works: protoWorks,
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

func (s *workService) RecoverWorks(ctx context.Context, req *personal_schedule.GetRecoveryWorksRequest) (*personal_schedule.GetRecoveryWorksResponse, error) {
	targetStart := time.Unix(0, req.TargetDate*int64(time.Millisecond))
	targetEnd := utils.EndOfDay(targetStart)
	sourceStart := time.Unix(0, req.SourceDate*int64(time.Millisecond))
	sourceEnd := utils.EndOfDay(sourceStart)
	timeShift := targetStart.Sub(sourceStart)

	targetStartMs := targetStart.UnixMilli()
	targetEndMs := targetEnd.UnixMilli()
	sourceStartMs := sourceStart.UnixMilli()
	sourceEndMs := sourceEnd.UnixMilli()

	draftLabels, err := s.workRepo.GetLabelsByTypeIDs(ctx, 6)
	if err != nil || len(draftLabels) == 0 {
		return nil, fmt.Errorf("system error: draft label not found")
	}
	draftLabelID := draftLabels[0].ID

	_ = s.workRepo.DeleteDraftsByDate(ctx, req.UserId, targetStart, targetEnd)

	sourceWorks, err := s.workRepo.GetAggregatedWorksByDateRangeMs(
		ctx,
		req.UserId,
		sourceStartMs,
		sourceEndMs,
	)
	if err != nil {
		s.logger.Error("Failed to get source works", "", zap.Error(err))
		return &personal_schedule.GetRecoveryWorksResponse{
			Error: utils.DatabaseError(ctx, err),
		}, nil
	}
	if len(sourceWorks) == 0 {
		return &personal_schedule.GetRecoveryWorksResponse{
			Works: []*personal_schedule.RecoveryWorkItem{},
		}, nil
	}

	targetWorks, err := s.workRepo.GetWorksByDateRangeMs(
		ctx,
		req.UserId,
		targetStartMs,
		targetEndMs,
	)

	if err != nil {
		s.logger.Error("Failed to get target works", "", zap.Error(err))
		return &personal_schedule.GetRecoveryWorksResponse{
			Error: utils.DatabaseError(ctx, err),
		}, nil
	}

	var responseItems []*personal_schedule.RecoveryWorkItem
	var worksToInsert []interface{}
	var subTasksToInsert []interface{}
	now := time.Now()

	for _, oldWork := range sourceWorks {
		newEnd := oldWork.EndDate.Add(timeShift)
		var newStart *time.Time
		hasStart := false
		if oldWork.StartDate != nil {
			t := oldWork.StartDate.Add(timeShift)
			newStart = &t
			hasStart = true
		}

		isConflict := false
		if hasStart {
			for _, existing := range targetWorks {
				var exStart time.Time
				if existing.StartDate != nil {
					exStart = *existing.StartDate
				}
				if newStart.Before(existing.EndDate) && newEnd.After(exStart) {
					isConflict = true
					break
				}
			}
		}

		newWorkID := bson.NewObjectID()

		oldSubTasks, err := s.workRepo.GetSubTasksByWorkID(ctx, oldWork.ID)
		if err != nil {
			s.logger.Error("Failed to get subtasks for work", "", zap.Error(err))
			return &personal_schedule.GetRecoveryWorksResponse{Error: utils.DatabaseError(ctx, err)}, nil
		}
		for _, oldSub := range oldSubTasks {
			nst := collection.SubTask{
				ID:             bson.NewObjectID(),
				Name:           oldSub.Name,
				IsCompleted:    false,
				WorkID:         newWorkID,
				CreatedAt:      now,
				LastModifiedAt: now,
			}
			subTasksToInsert = append(subTasksToInsert, nst)

		}

		var goalID *bson.ObjectID
		if len(oldWork.GoalInfo) > 0 {
			gid := oldWork.GoalInfo[0].ID
			goalID = &gid
		}

		fmt.Println("Old Work ID:", oldWork.ID.Hex(), "New Work ID:", newWorkID.Hex())
		newWorkDB := collection.Work{
			ID:                  newWorkID,
			Name:                oldWork.Name,
			ShortDescriptions:   oldWork.ShortDescriptions,
			DetailedDescription: oldWork.DetailedDescription,
			StartDate:           newStart,
			EndDate:             newEnd,
			UserID:              oldWork.UserID,
			StatusID:            oldWork.Status[0].ID,
			DifficultyID:        oldWork.Difficulty[0].ID,
			PriorityID:          oldWork.Priority[0].ID,
			TypeID:              oldWork.Type[0].ID,
			CategoryID:          oldWork.Category[0].ID,
			GoalID:              goalID,
			DraftID:             &draftLabelID,
			CreatedAt:           now,
			LastModifiedAt:      now,
		}
		worksToInsert = append(worksToInsert, newWorkDB)

		tempAggWork := oldWork
		tempAggWork.ID = newWorkID
		tempAggWork.StartDate = newStart
		tempAggWork.EndDate = newEnd

		protoWorks := s.workMapper.ConvertAggregatedWorksToProto([]repos.AggregatedWork{tempAggWork})
		workDetailProto := protoWorks[0]
		workDetailProto.Id = ""

		responseItems = append(responseItems, &personal_schedule.RecoveryWorkItem{
			Work:       workDetailProto,
			IsConflict: isConflict,
		})
	}

	if len(worksToInsert) > 0 {
		if err := s.workRepo.BulkInsertWorks(ctx, worksToInsert); err != nil {
			s.logger.Error("Failed to bulk insert recovered works", "", zap.Error(err))
			return &personal_schedule.GetRecoveryWorksResponse{Error: utils.DatabaseError(ctx, err)}, nil
		}
	}
	if len(subTasksToInsert) > 0 {
		s.workRepo.BulkInsertSubTasks(ctx, subTasksToInsert)
	}

	return &personal_schedule.GetRecoveryWorksResponse{
		Works: responseItems,
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
