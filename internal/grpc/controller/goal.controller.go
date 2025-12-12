package controller

import (
	"context"
	"personal_schedule_service/internal/grpc/services"
	"personal_schedule_service/internal/grpc/utils"
	"personal_schedule_service/proto/personal_schedule"
)

type GoalController struct {
	personal_schedule.UnimplementedGoalServiceServer
	goalService services.GoalService
}

func NewGoalController(
	goalService services.GoalService,
) *GoalController {
	return &GoalController{
		goalService: goalService,
	}
}

func (g *GoalController) GetGoals(ctx context.Context, req *personal_schedule.GetGoalsRequest) (*personal_schedule.GetGoalsResponse, error) {
	return utils.WithSafePanic(ctx, req, g.goalService.GetGoals)
}

func (g *GoalController) UpsertGoal(ctx context.Context, req *personal_schedule.UpsertGoalRequest) (*personal_schedule.UpsertGoalResponse, error) {
	return utils.WithSafePanic(ctx, req, g.goalService.UpsertGoal)
}

func (g *GoalController) GetGoal(ctx context.Context, req *personal_schedule.GetGoalRequest) (*personal_schedule.GetGoalResponse, error) {
	return utils.WithSafePanic(ctx, req, g.goalService.GetGoal)
}

func (g *GoalController) DeleteGoal(ctx context.Context, req *personal_schedule.DeleteGoalRequest) (*personal_schedule.DeleteGoalResponse, error) {
	return utils.WithSafePanic(ctx, req, g.goalService.DeleteGoal)
}

func (g *GoalController) GetGoalForDiaglog(ctx context.Context, req *personal_schedule.GetGoalsForDialogRequest) (*personal_schedule.GetGoalForDialogResponse, error) {
	return utils.WithSafePanic(ctx, req, g.goalService.GetGoalsForDialog)
}

func (g *GoalController) UpdateGoalLabel(ctx context.Context, req *personal_schedule.UpdateGoalLabelRequest) (*personal_schedule.UpdateGoalLabelResponse, error) {
	return utils.WithSafePanic(ctx, req, g.goalService.UpdateGoalLabel)
}
