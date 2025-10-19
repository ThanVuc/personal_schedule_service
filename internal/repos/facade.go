package repos

import (
	"context"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/internal/grpc/models"
)

type (
	UserRepo interface {
		UpsertSyncUser(ctx context.Context, payload models.UserOutboxPayload, requestId string) error
	}

	LabelRepo interface {
		CountLabels(ctx context.Context) (int, error)
		InsertLabels(ctx context.Context, labels *[]collection.Label) error
		GetLabels(ctx context.Context) ([]collection.Label, error)
		GetLabelsByTypeIDs(ctx context.Context, typeIDs int32) ([]collection.Label, error)
	}
)

func NewUserRepo() UserRepo {
	return &userRepo{
		logger:    global.Logger,
		connector: global.MongoDbConntector,
	}
}

func NewLabelRepo() LabelRepo {
	return &labelRepo{
		logger:         global.Logger,
		mongoConnector: global.MongoDbConntector,
	}
}
