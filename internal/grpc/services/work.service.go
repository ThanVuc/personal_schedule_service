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

	// --- BƯỚC 3: XỬ LÝ WORK CHÍNH (Đã được đơn giản hóa) ---
	if req.Id == nil || *req.Id == "" {
		// === TẠO MỚI (CREATE) WORK ===
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

	// --- BƯỚC 5: TRẢ VỀ ---
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
		task.WorkID = workID // Gán WorkID

		if task.ID.IsZero() {
			// CREATE
			task.ID = bson.NewObjectID()
			task.CreatedAt = now
			task.LastModifiedAt = now
			operations = append(operations, mongo.NewInsertOneModel().SetDocument(task))
		} else {
			// UPDATE
			delete(existingTaskMap, task.ID)
			operations = append(operations, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": task.ID, "work_id": workID}). // Filter an toàn
				SetUpdate(bson.M{"$set": bson.M{
					"name":             task.Name,
					"is_completed":     task.IsCompleted,
					"last_modified_at": now,
				}}))
		}
	}

	// DELETE
	for taskID := range existingTaskMap {
		operations = append(operations, mongo.NewDeleteOneModel().SetFilter(bson.M{"_id": taskID, "work_id": workID}))
	}

	if len(operations) > 0 {
		_, err = s.workRepo.BulkWriteSubTasks(ctx, operations)
		return err
	}

	return nil
}
