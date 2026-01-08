package services

import (
	"context"
	"fmt"

	// "fmt"
	"personal_schedule_service/internal/collection"
	labels_constant "personal_schedule_service/internal/constant/labels"
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

type goalService struct {
	logger         log.Logger
	goalRepo       repos.GoalRepo
	goalMapper     mapper.GoalMapper
	validator      validation.GoalValidator
	mongoConnector *mongolib.MongoConnector
}

func (s *goalService) GetGoals(ctx context.Context, req *personal_schedule.GetGoalsRequest) (*personal_schedule.GetGoalsResponse, error) {
	goals, totalGoals, err := s.goalRepo.GetGoals(ctx, req)
	if err != nil {
		s.logger.Error("Error fetching goals from repo", "err", zap.Error(err))
		return &personal_schedule.GetGoalsResponse{
			Error:    utils.InternalServerError(ctx, err),
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

	overdueLabel, err := s.goalRepo.GetLabelByKey(ctx, labels_constant.LabelOverDue)
	if err != nil {
		s.logger.Error("Failed to get overdue label", "", zap.Error(err))
	}
	timeNow := time.Now()
	for i, goal := range goals {
		if goal.EndDate != nil && goal.EndDate.Before(timeNow) {
			goals[i].Overdue = []collection.Label{*overdueLabel}
		}
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
	requestID := utils.GetRequestIDFromOutgoingContext(ctx)
	if err := s.validator.ValidationGoal(ctx, req); err != nil {
		s.logger.Warn("Goal validation failed", "", zap.String("request_id", requestID), zap.Error(err))
		if ve, ok := err.(*validation.ValidationError); ok {
			return &personal_schedule.UpsertGoalResponse{
				IsSuccess: false,
				Error:     utils.CustomError(ctx, ve.Category, ve.Code, err),
			}, nil
		}
	}

	goalDB, tasksDB, err := s.goalMapper.MapUpsertProtoToModels(req)
	if err != nil {
		s.logger.Warn("Failed to map UpsertGoal proto", "", zap.Error(err))
		errMsg := "Invalid data format: " + err.Error()
		return &personal_schedule.UpsertGoalResponse{
			IsSuccess: false,
			Message:   errMsg,
			Error:     utils.InternalServerError(ctx, err),
		}, nil
	}

	var goalID bson.ObjectID
	now := time.Now()

	if req.Id == nil || *req.Id == "" {
		goalDB.UserID = req.UserId
		goalDB.CreatedAt = now
		goalDB.LastModifiedAt = now

		newID, err := s.goalRepo.CreateGoal(ctx, goalDB)
		if err != nil {
			s.logger.Error("Failed to create goal", "", zap.Error(err))
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
			s.logger.Error("Failed to get goal by ID", "", zap.Error(err))
			return &personal_schedule.UpsertGoalResponse{
				IsSuccess: false,
				Message:   "Database error",
				Error:     utils.DatabaseError(ctx, err),
			}, err
		}
		if existingGoal == nil {
			errMsg := "goal not found"
			s.logger.Warn(errMsg, "", zap.String("goal_id", *req.Id))
			return &personal_schedule.UpsertGoalResponse{
				IsSuccess: false,
				Message:   errMsg,
				Error:     utils.NotFoundError(ctx, err),
			}, nil
		}
		if existingGoal.UserID != req.UserId {
			errMsg := "forbidden: user does not own this goal"
			s.logger.Warn(errMsg, "", zap.String("goal_id", *req.Id), zap.String("user_id", req.UserId))
			return &personal_schedule.UpsertGoalResponse{
				IsSuccess: false,
				Message:   errMsg,
				Error:     utils.PermissionDeniedError(ctx, err),
			}, nil
		}

		if err := s.goalRepo.UpdateGoal(ctx, goalID, goalDB); err != nil {
			s.logger.Error("Failed to update goal", "", zap.Error(err))
			return &personal_schedule.UpsertGoalResponse{
				IsSuccess: false,
				Message:   "Failed to update goal",
				Error:     utils.DatabaseError(ctx, err),
			}, err
		}
	}

	if err := s.syncGoalTasks(ctx, goalID, tasksDB); err != nil {
		s.logger.Error("Failed to sync goal tasks", "", zap.Error(err))
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

func (s *goalService) GetGoal(ctx context.Context, req *personal_schedule.GetGoalRequest) (*personal_schedule.GetGoalResponse, error) {
	goalID, err := bson.ObjectIDFromHex(req.GoalId)
	if err != nil {
		s.logger.Warn("Invalid goal ID format", "", zap.String("goal_id", req.GoalId), zap.Error(err))
		return &personal_schedule.GetGoalResponse{
			Goal:  nil,
			Error: utils.InternalServerError(ctx, err),
		}, nil
	}

	goal, err := s.goalRepo.GetAggregatedGoalByID(ctx, goalID)
	if err != nil {
		s.logger.Error("Error fetching goal from repo", "err", zap.Error(err))
		return &personal_schedule.GetGoalResponse{
			Goal:  nil,
			Error: utils.DatabaseError(ctx, err),
		}, err
	}

	if goal == nil {
		s.logger.Info("Goal not found", "", zap.String("goal_id", req.GoalId))
		return &personal_schedule.GetGoalResponse{
			Goal:  nil,
			Error: utils.NotFoundError(ctx, err),
		}, nil
	}

	if goal.UserID != req.UserId {
		s.logger.Warn("Forbidden: user does not own this goal", "", zap.String("goal_id", req.GoalId), zap.String("user_id", req.UserId))
		return &personal_schedule.GetGoalResponse{
			Goal:  nil,
			Error: utils.PermissionDeniedError(ctx, err),
		}, nil
	}

	taskDB, err := s.goalRepo.GetTasksByGoalID(ctx, goalID)
	if err != nil {
		s.logger.Error("Error fetching goal tasks from repo", "err", zap.Error(err))
		return &personal_schedule.GetGoalResponse{
			Goal:  nil,
			Error: utils.DatabaseError(ctx, err),
		}, err
	}

	protoGoal := s.goalMapper.MapAggregatedToDetailProto(*goal, taskDB)

	return &personal_schedule.GetGoalResponse{
		Goal:  protoGoal,
		Error: nil,
	}, nil
}

func (s *goalService) DeleteGoal(ctx context.Context, req *personal_schedule.DeleteGoalRequest) (*personal_schedule.DeleteGoalResponse, error) {
	goalID, err := bson.ObjectIDFromHex(req.GoalId)
	if err != nil {
		s.logger.Warn("Invalid goal ID format", "", zap.String("goal_id", req.GoalId), zap.Error(err))
		return &personal_schedule.DeleteGoalResponse{
			Success: false,
			Error:   utils.InternalServerError(ctx, err),
		}, nil
	}

	goal, err := s.goalRepo.GetGoalByID(ctx, goalID)
	if err != nil {
		s.logger.Error("Error fetching goal from repo", "err", zap.Error(err))
		return nil, err
	}
	if goal == nil {
		s.logger.Info("Goal not found", "", zap.String("goal_id", req.GoalId))
		return nil, nil
	}

	if goal.UserID != req.UserId {
		return nil, fmt.Errorf("forbidden: user does not own this goal")
	}

	if err := s.goalRepo.DeleteTasksByGoalID(ctx, goalID); err != nil {
		s.logger.Error("Error deleting goal tasks", "err", zap.Error(err))
		return nil, err
	}
	if err := s.goalRepo.DeleteGoal(ctx, goalID); err != nil {
		s.logger.Error("Error deleting goal", "err", zap.Error(err))
		return nil, err
	}

	return &personal_schedule.DeleteGoalResponse{
		Success: true,
		Error:   nil,
	}, nil
}

func (s *goalService) GetGoalsForDialog(ctx context.Context, req *personal_schedule.GetGoalsForDialogRequest) (*personal_schedule.GetGoalForDialogResponse, error) {
	goals, err := s.goalRepo.GetGoalsForDialog(ctx, req.UserId)
	if err != nil {
		s.logger.Error("Failed to get goals for dialog", "", zap.Error(err))
		return &personal_schedule.GetGoalForDialogResponse{}, err
	}

	respItems := make([]*personal_schedule.GoalOfWork, len(goals))
	for i, g := range goals {
		respItems[i] = &personal_schedule.GoalOfWork{
			Id:   g.ID.Hex(),
			Name: g.Name,
		}
	}

	return &personal_schedule.GetGoalForDialogResponse{
		Goals: respItems,
	}, nil
}

func (s *goalService) UpdateGoalLabel(ctx context.Context, req *personal_schedule.UpdateGoalLabelRequest) (*personal_schedule.UpdateGoalLabelResponse, error) {
	goalID, err := bson.ObjectIDFromHex(req.GoalId)
	if err != nil {
		s.logger.Warn("Invalid goal ID format", "", zap.String("goal_id", req.GoalId), zap.Error(err))
		return &personal_schedule.UpdateGoalLabelResponse{
			Error: utils.InternalServerError(ctx, err),
		}, nil
	}

	labelID, err := bson.ObjectIDFromHex(req.LabelId)
	if err != nil {
		s.logger.Warn("Invalid label ID format", "", zap.String("label_id", req.LabelId), zap.Error(err))
		return &personal_schedule.UpdateGoalLabelResponse{
			Error: utils.InternalServerError(ctx, err),
		}, nil
	}

	goal, err := s.goalRepo.GetGoalByID(ctx, goalID)
	if err != nil {
		s.logger.Error("Failed to get goal by ID", "", zap.Error(err))
		return &personal_schedule.UpdateGoalLabelResponse{
			Error: utils.DatabaseError(ctx, err),
		}, nil
	}

	if goal == nil {
		errMsg := "goal not found"
		s.logger.Warn(errMsg, "", zap.String("goal_id", req.GoalId))
		return &personal_schedule.UpdateGoalLabelResponse{
			Error: utils.NotFoundError(ctx, err),
		}, nil
	}
	if goal.UserID != req.UserId {
		errMsg := "forbidden: user does not own this goal"
		s.logger.Warn(errMsg, "", zap.String("goal_id", req.GoalId), zap.String("user_id", req.UserId))
		return &personal_schedule.UpdateGoalLabelResponse{
			Error: utils.PermissionDeniedError(ctx, err),
		}, nil
	}

	var fieldName string
	switch req.LabelType {
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
		return &personal_schedule.UpdateGoalLabelResponse{
			Error: utils.InternalServerError(ctx, err),
		}, nil
	}
	if err := s.goalRepo.UpdateGoalField(ctx, goalID, fieldName, labelID); err != nil {
		s.logger.Error("Failed to update goal field", "", zap.Error(err))
		return &personal_schedule.UpdateGoalLabelResponse{
			Error: utils.DatabaseError(ctx, err),
		}, nil
	}

	return &personal_schedule.UpdateGoalLabelResponse{
		Error:   nil,
		Message: "Goal label updated successfully",
	}, nil
}
