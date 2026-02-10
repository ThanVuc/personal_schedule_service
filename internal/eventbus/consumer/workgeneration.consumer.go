package consumer

import (
	"context"
	"personal_schedule_service/global"
	workgeneration_constant "personal_schedule_service/internal/constant/work"
	"personal_schedule_service/internal/eventbus/handler"

	"github.com/thanvuc/go-core-lib/eventbus"
	"github.com/thanvuc/go-core-lib/log"
	"github.com/wagslane/go-rabbitmq"
)

type WorkGenerationConsumer struct {
	logger  log.Logger
	handler *handler.WorkGenerationHandler
}

func NewWorkGenerationConsumer(
	handler *handler.WorkGenerationHandler,
) *WorkGenerationConsumer {
	return &WorkGenerationConsumer{
		logger:  global.Logger,
		handler: handler,
	}
}

func (c *WorkGenerationConsumer) ConsumeWorks(ctx context.Context) {
	workConsumer := eventbus.NewConsumer(
		global.EventBusConnector,
		workgeneration_constant.WORK_TRANSFER_EXCHANGE,
		eventbus.ExchangeTypeDirect,
		workgeneration_constant.WORK_TRANSFER_ROUTING_KEY,
		workgeneration_constant.WORK_TRANSFER_QUEUE,
		1,
	)

	c.logger.Info("Starting to consume messages from consumer works", "")
	go func() {
		err := workConsumer.Consume(ctx, func(d rabbitmq.Delivery) (action rabbitmq.Action) {
			return c.handler.ConsumeWorks(ctx, d)
		})

		if err != nil {
			c.logger.Error("Failed to consume messages from consumer works", "")
			return
		}
	}()
}
