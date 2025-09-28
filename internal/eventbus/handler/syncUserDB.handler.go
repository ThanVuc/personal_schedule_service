package handler

import (
	"context"
	"encoding/json"
	"personal_schedule_service/global"

	"github.com/thanvuc/go-core-lib/log"
)

type SyncAuthHandler struct {
	logger log.Logger
}

type UserOutboxPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	CreatedAt int64  `json:"created_at"`
}

func NewSyncAuthHandler() *SyncAuthHandler {
	return &SyncAuthHandler{
		logger: global.Logger,
	}
}

func (h *SyncAuthHandler) SyncUserDB(ctx context.Context, payload []byte) error {
	var userPayload UserOutboxPayload
	if err := json.Unmarshal(payload, &userPayload); err != nil {
		return err
	}

	return nil
}
