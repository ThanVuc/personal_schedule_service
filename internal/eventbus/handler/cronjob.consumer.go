package handler

import (
	"context"

	"github.com/wagslane/go-rabbitmq"
)

type CronJobHandler struct {
}

func (h *CronJobHandler) SetGiveupWorkLabel(ctx context.Context, d rabbitmq.Delivery) rabbitmq.Action {
	return rabbitmq.Ack
}

func (h *CronJobHandler) CreateTodayRepeatedWorks(ctx context.Context, d rabbitmq.Delivery) rabbitmq.Action {
	return rabbitmq.Ack
}
