package controller

import (
	"context"
	"personal_schedule_service/internal/grpc/services"
	"personal_schedule_service/internal/grpc/utils"
	"personal_schedule_service/proto/common"
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

func (wc *WorkController) DeleteWork(ctx context.Context, req *personal_schedule.DeleteWorkRequest) (*personal_schedule.DeleteWorkResponse, error) {
	return utils.WithSafePanic(ctx, req, wc.workService.DeleteWork)
}

func (wc *WorkController) GetRecoveryWorks(ctx context.Context, req *personal_schedule.GetRecoveryWorksRequest) (*personal_schedule.GetRecoveryWorksResponse, error) {
	return utils.WithSafePanic(ctx, req, wc.workService.RecoverWorks)
}

func (wc *WorkController) UpdateWorkLabel(ctx context.Context, req *personal_schedule.UpdateWorkLabelRequest) (*personal_schedule.UpdateWorkLabelResponse, error) {
	return utils.WithSafePanic(ctx, req, wc.workService.UpdateWorkLabel)
}

func (wc *WorkController) SaveDraftAsRealWork(ctx context.Context, req *personal_schedule.SaveDraftAsRealWorkRequest) (*personal_schedule.SaveDraftAsRealWorkResponse, error) {
	return utils.WithSafePanic(ctx, req, wc.workService.SaveDraftAsRealWork)
}

func (wc *WorkController) DeleteAllDraftWorks(ctx context.Context, req *personal_schedule.DeleteAllDraftWorksRequest) (*personal_schedule.DeleteAllDraftWorksResponse, error) {
	return utils.WithSafePanic(ctx, req, wc.workService.DeleteAllDraftWorks)
}

func (wc *WorkController) GenerateWorksByAI(ctx context.Context, req *personal_schedule.GenerateWorksByAIRequest) (*common.EmptyResponse, error) {
	return utils.WithSafePanic(ctx, req, wc.workService.GenerateWorksFromAI)
}
