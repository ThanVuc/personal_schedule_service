package repos

import (
	"context"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/grpc/models"
)

type (
	UserRepo interface {
		UpsertSyncUser(ctx context.Context, payload models.UserOutboxPayload, requestId string) error
	}
)

func NewUserRepo() UserRepo {
	return &userRepo{
		logger:    global.Logger,
		connector: global.MongoDbConntector,
	}
}
