package services

import (
	"context"
	// "fmt"
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/internal/grpc/mapper"
	"personal_schedule_service/internal/grpc/utils"
	"personal_schedule_service/internal/repos"
	"personal_schedule_service/proto/personal_schedule"
	"time"

	"github.com/thanvuc/go-core-lib/log"
	"github.com/thanvuc/go-core-lib/mongolib"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

type goalService struct {
	logger         log.Logger
	goalRepo       repos.GoalRepo
	goalMapper     mapper.GoalMapper
	mongoConnector *mongolib.MongoConnector
}

func (s *goalService) GetGoals(ctx context.Context, req *personal_schedule.GetGoalsRequest) (*personal_schedule.GetGoalsResponse, error) {
	goals, totalGoals, err := s.goalRepo.GetGoals(ctx, req)

	if err != nil {
		s.logger.Error("Error fetching goals from repo", "err", zap.Error(err))
		return &personal_schedule.GetGoalsResponse{
			Error:    utils.DatabaseError(ctx, err),
			Goals:    nil,
			PageInfo: utils.ToPageInfo(req.PageQuery.Page, req.PageQuery.PageSize, int32(totalGoals)),
		}, err
	}

	if totalGoals == 0 {
		return &personal_schedule.GetGoalsResponse{
			Error:      nil,
			Goals:      nil,
			TotalGoals: 0,
			PageInfo:   utils.ToPageInfo(req.PageQuery.Page, req.PageQuery.PageSize, int32(totalGoals)),
		}, nil
	}

	protoGoals := s.goalMapper.ConvertAggregatedGoalsToProto(goals)

	pageInfo := utils.ToPageInfo(req.PageQuery.Page, req.PageQuery.PageSize, int32(totalGoals))

	resp := &personal_schedule.GetGoalsResponse{
		Goals:      protoGoals,
		PageInfo:   pageInfo,
		TotalGoals: int32(totalGoals),
		Error:      nil,
	}

	return resp, nil
}

func (s *goalService) UpsertGoal(ctx context.Context, req *personal_schedule.UpsertGoalRequest) (*personal_schedule.UpsertGoalResponse, error) {
	goalDB, tasksDB, err := s.goalMapper.MapUpsertProtoToModels(req)
	if err != nil {
		s.logger.Warn("Failed to map UpsertGoal proto", "",zap.Error(err))
		errMsg := "Invalid data format: " + err.Error()
		return &personal_schedule.UpsertGoalResponse{
			IsSuccess: false,
			Message:   errMsg,
			Error:     utils.InternalServerError(ctx, err),
		}, nil
	}

	var goalID bson.ObjectID
	now := time.Now()

	if req.Id == nil || *req.Id == ""  { 
		goalDB.UserID = req.UserId
		goalDB.CreatedAt = now
		goalDB.LastModifiedAt = now

		newID, err := s.goalRepo.CreateGoal(ctx, goalDB)
		if err != nil {
			s.logger.Error("Failed to create goal", "",zap.Error(err))
			return &personal_schedule.UpsertGoalResponse{
				IsSuccess: false,
				Message:   "Failed to create goal",
				Error:     utils.DatabaseError(ctx, err),
			}, err
		}
		goalID = newID

	} else {
		goalID, _ = bson.ObjectIDFromHex(*req.Id)

		existingGoal, err := s.goalRepo.GetGoalByID(ctx, goalID)
		if err != nil {
			s.logger.Error("Failed to get goal by ID", "",zap.Error(err))
			return &personal_schedule.UpsertGoalResponse{
				IsSuccess: false,
				Message:   "Database error",
				Error:     utils.DatabaseError(ctx, err),
			}, err
		}
		if existingGoal == nil {
			errMsg := "goal not found"
			s.logger.Warn(errMsg,"", zap.String("goal_id", *req.Id))
			return &personal_schedule.UpsertGoalResponse{
				IsSuccess: false,
				Message:   errMsg,
				Error:     utils.NotFoundError(ctx, err ),
			}, nil
		}
		if existingGoal.UserID != req.UserId {
			errMsg := "forbidden: user does not own this goal"
			s.logger.Warn(errMsg, "",zap.String("goal_id", *req.Id), zap.String("user_id", req.UserId))
			return &personal_schedule.UpsertGoalResponse{
				IsSuccess: false,
				Message:   errMsg,
				Error:     utils.PermissionDeniedError(ctx, err),
			}, nil
		}

		updates := bson.M{
			"name":                 goalDB.Name,
			"short_descriptions":   goalDB.ShortDescriptions,
			"detailed_description": goalDB.DetailedDescription,
			"start_date":           goalDB.StartDate,
			"end_date":             goalDB.EndDate,
			"status_id":            goalDB.StatusID,
			"difficulty_id":        goalDB.DifficultyID,
			"priority_id":          goalDB.PriorityID,
			"last_modified_at":     now,
		}

		if err := s.goalRepo.UpdateGoal(ctx, goalID, updates); err != nil {
			s.logger.Error("Failed to update goal","", zap.Error(err))
			return &personal_schedule.UpsertGoalResponse{
				IsSuccess: false,
				Message:   "Failed to update goal",
				Error:     utils.DatabaseError(ctx, err),
			}, err
		}
	}

	if err := s.syncGoalTasks(ctx, goalID, tasksDB); err != nil {
		s.logger.Error("Failed to sync goal tasks","", zap.Error(err))
		return &personal_schedule.UpsertGoalResponse{
			IsSuccess: false,
			Message:   "Failed to sync tasks (Goal was upserted but tasks failed)",
			Error:     utils.DatabaseError(ctx, err),
		}, err
	}

	return &personal_schedule.UpsertGoalResponse{
		IsSuccess: true,
		Message:   "Goal upserted successfully",
	}, nil
}


func (s *goalService) syncGoalTasks(ctx context.Context, goalID bson.ObjectID, payloadTasks []collection.GoalTask) error {
	existingTasks, err := s.goalRepo.GetTasksByGoalID(ctx, goalID)
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
		task.GoalID = goalID

		if task.ID.IsZero() {
			task.ID = bson.NewObjectID()
			task.CreatedAt = now
			task.LastModifiedAt = now
			operations = append(operations, mongo.NewInsertOneModel().SetDocument(task))
		} else {
			delete(existingTaskMap, task.ID)
			operations = append(operations, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": task.ID, "goal_id": goalID}).
				SetUpdate(bson.M{"$set": bson.M{
					"name":             task.Name,
					"is_completed":     task.IsCompleted,
					"last_modified_at": now,
				}}))
		}
	}

	for taskID := range existingTaskMap {
		operations = append(operations, mongo.NewDeleteOneModel().SetFilter(bson.M{"_id": taskID, "goal_id": goalID}))
	}

	if len(operations) > 0 {
		_, err = s.goalRepo.BulkWriteTasks(ctx, operations)
		return err
	}

	return nil
}
