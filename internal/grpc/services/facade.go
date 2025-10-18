package services

import (
	"context"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/grpc/helper"
	"personal_schedule_service/internal/grpc/mapper"
	"personal_schedule_service/internal/repos"
	"personal_schedule_service/proto/common"
	"personal_schedule_service/proto/personal_schedule"
)

type (
	LabelService interface {
		SeedLabels(ctx context.Context) error
		GetLabelPerTypes(ctx context.Context, req *common.EmptyRequest) (*personal_schedule.GetLabelPerTypesResponse, error)
	}
)

func NewLabelService(
	labelRepo repos.LabelRepo,
	labelMapper mapper.LabelMapper,
) LabelService {
	return &labelService{
		labelHelper: helper.NewLabelHelper(),
		logger:      global.Logger,
		labelRepo:   labelRepo,
		labelMapper: labelMapper,
	}
}
