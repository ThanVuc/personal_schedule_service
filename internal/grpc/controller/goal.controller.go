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
