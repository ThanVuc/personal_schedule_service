package controller

import (
	"context"
	"personal_schedule_service/internal/grpc/services"
	"personal_schedule_service/internal/grpc/utils"
	"personal_schedule_service/proto/common"
	"personal_schedule_service/proto/personal_schedule"
)

type LabelController struct {
	personal_schedule.UnimplementedLabelServiceServer
	labelService services.LabelService
}

func NewLabelController(
	labelService services.LabelService,
) *LabelController {
	return &LabelController{
		labelService: labelService,
	}
}

func (lc *LabelController) GetLabelPerTypes(ctx context.Context, req *common.EmptyRequest) (*personal_schedule.GetLabelPerTypesResponse, error) {
	return utils.WithSafePanic(ctx, req, lc.labelService.GetLabelPerTypes)
}

func (lc *LabelController) GetLabelsByTypeIDs(ctx context.Context, req *common.IDRequest) (*personal_schedule.GetLabelsByTypeIDsResponse, error) {
	return utils.WithSafePanic(ctx, req, lc.labelService.GetLabelsByTypeIDs)
}
