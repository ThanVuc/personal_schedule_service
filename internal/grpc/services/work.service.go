package services

import (
	"context"
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/internal/grpc/mapper"
	"personal_schedule_service/internal/grpc/utils"
	"personal_schedule_service/internal/grpc/validation"
	"personal_schedule_service/internal/repos"
	"personal_schedule_service/proto/personal_schedule"
	"time"

	"github.com/thanvuc/go-core-lib/log"
	"github.com/thanvuc/go-core-lib/mongolib"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

type workService struct {
	logger         log.Logger
	workRepo       repos.WorkRepo
	workMapper     mapper.WorkMapper
	mongoConnector *mongolib.MongoConnector
	validator      validation.WorkValidator
}

func (s *workService) UpsertWork(ctx context.Context, req *personal_schedule.UpsertWorkRequest) (*personal_schedule.UpsertWorkResponse, error) {
	if err := s.validator.ValidateUpsertWork(ctx, req); err != nil {
		s.logger.Warn("UpsertWork validation failed", "", zap.Error(err))
		return &personal_schedule.UpsertWorkResponse{
			IsSuccess: false,
			Error:     utils.InternalServerError(ctx, err),
		}, nil
	}

	workDB, subTasksDB, err := s.workMapper.MapUpsertProtoToModels(req)
	if err != nil {
		s.logger.Error("Failed to map UpsertWorkRequest to models", "", zap.Error(err))
		return &personal_schedule.UpsertWorkResponse{
			IsSuccess: false,
			Error:     utils.InternalServerError(ctx, err),
		}, nil
	}

	var workID bson.ObjectID
	now := time.Now()
	if req.Id == nil || *req.Id == "" {
		workDB.UserID = req.UserId
		workDB.CreatedAt = now
		workDB.LastModifiedAt = now

		newID, err := s.workRepo.CreateWork(ctx, workDB)
		if err != nil {
			s.logger.Error("Failed to create work", "", zap.Error(err))
			return &personal_schedule.UpsertWorkResponse{
				IsSuccess: false, Message: "Failed to create work", Error: utils.DatabaseError(ctx, err),
			}, err
		}
		workID = newID

	} else {
		workID, _ = bson.ObjectIDFromHex(*req.Id)

		updates := bson.M{
			"name":                 workDB.Name,
			"short_descriptions":   workDB.ShortDescriptions,
			"detailed_description": workDB.DetailedDescription,
			"start_date":           workDB.StartDate,
			"end_date":             workDB.EndDate,
			"notification_ids":     workDB.NotificationIds,
			"status_id":            workDB.StatusID,
			"difficulty_id":        workDB.DifficultyID,
			"priority_id":          workDB.PriorityID,
			"type_id":              workDB.TypeID,
			"category_id":          workDB.CategoryID,
			"goal_id":              workDB.GoalID,
			"last_modified_at":     now,
		}

		if err := s.workRepo.UpdateWork(ctx, workID, updates); err != nil {
			s.logger.Error("Failed to update work", "", zap.Error(err))
			return &personal_schedule.UpsertWorkResponse{
				IsSuccess: false, Message: "Failed to update work", Error: utils.DatabaseError(ctx, err),
			}, err
		}
	}
	if err := s.syncSubTasks(ctx, workID, subTasksDB); err != nil {
		s.logger.Error("Failed to sync sub-tasks", "", zap.Error(err))
		return &personal_schedule.UpsertWorkResponse{
			IsSuccess: false, Message: "Failed to sync sub-tasks (Work was upserted but tasks failed)", Error: utils.DatabaseError(ctx, err),
		}, err
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
