package handler

import (
	"context"
	"encoding/json"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/grpc/models"
	"personal_schedule_service/internal/repos"

	"github.com/thanvuc/go-core-lib/log"
)

type SyncAuthHandler struct {
	logger   log.Logger
	userRepo repos.UserRepo
}

func NewSyncAuthHandler(
	userRepo repos.UserRepo,
) *SyncAuthHandler {
	return &SyncAuthHandler{
		logger:   global.Logger,
		userRepo: userRepo,
	}
}

func (h *SyncAuthHandler) SyncUserDB(ctx context.Context, payload []byte, requestId string) error {
	var userPayload models.UserOutboxPayload
	if err := json.Unmarshal(payload, &userPayload); err != nil {
		return err
	}

	if err := h.userRepo.UpsertSyncUser(ctx, userPayload, requestId); err != nil {
		return err
	}

	return nil
}
