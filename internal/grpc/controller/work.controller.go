package controller

import (
	"context"
	"personal_schedule_service/internal/grpc/services"
	"personal_schedule_service/internal/grpc/utils"
	"personal_schedule_service/proto/personal_schedule"
)

type WorkController struct {
	personal_schedule.UnimplementedWorkServiceServer
	workService services.WorkService
}

func NewWorkController(
	WorkService services.WorkService,
) *WorkController {
	return &WorkController{
		workService: WorkService,
	}
}

func (wc *WorkController) UpsertWork(ctx context.Context, req *personal_schedule.UpsertWorkRequest) (*personal_schedule.UpsertWorkResponse, error) {
	return utils.WithSafePanic(ctx, req, wc.workService.UpsertWork)
}

func (wc *WorkController) GetWorks(ctx context.Context, req *personal_schedule.GetWorksRequest) (*personal_schedule.GetWorksResponse, error) {
	return utils.WithSafePanic(ctx, req, wc.workService.GetWorks)
}

func (wc *WorkController) GetWork(ctx context.Context, req *personal_schedule.GetWorkRequest) (*personal_schedule.GetWorkResponse, error) {
	return utils.WithSafePanic(ctx, req, wc.workService.GetWork)
}
