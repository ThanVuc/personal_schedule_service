package services

import (
	"context"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/grpc/helper"
)

type (
	LabelService interface {
		SeedLabels(ctx context.Context) error
	}
)

func NewLabelService() LabelService {
	return &labelService{
		mongoConnector: *global.MongoDbConntector,
		labelHelper:    &helper.LabelHelper{},
		logger:         global.Logger,
	}
}
