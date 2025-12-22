package consumer

import (
	"context"
	"personal_schedule_service/global"
	eventbus_constant "personal_schedule_service/internal/eventbus/constant"
	"personal_schedule_service/internal/eventbus/handler"

	"github.com/thanvuc/go-core-lib/eventbus"
	"github.com/thanvuc/go-core-lib/log"
	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"
)

type CronJobConsumer struct {
	logger         log.Logger
	cronJobHandler *handler.CronJobHandler
}

func NewCronJobConsumer(logger log.Logger) *CronJobConsumer {
	return &CronJobConsumer{
		logger:         logger,
		cronJobHandler: &handler.CronJobHandler{},
	}
}

func (c *CronJobConsumer) ConsumeGiveUpJob(ctx context.Context) {
	c.logger.Info("Starting to consume give up job messages", "")
	consumer := eventbus.NewConsumer(
		global.EventBusConnector,
		eventbus_constant.VIET_NAM_JOB_EXCHANGE,
		eventbus.ExchangeTypeTopic,
		eventbus_constant.ONE_DAY_JOB_ROUTING_KEY,
		eventbus_constant.GIVE_UP_JOB_QUEUE,
		1,
	)

	go func() {
		err := consumer.Consume(ctx, func(d rabbitmq.Delivery) (action rabbitmq.Action) {
			return c.cronJobHandler.SetGiveupWorkLabel(ctx, d)

		})

		if err != nil {
			c.logger.Error("Failed to consume give up job messages", "", zap.Error(err))
			return
		}
	}()
}

func (c *CronJobConsumer) ConsumeCreateTodayRepeatedWorksJob(ctx context.Context) {
	c.logger.Info("Starting to consume give up job messages", "")
	consumer := eventbus.NewConsumer(
		global.EventBusConnector,
		eventbus_constant.VIET_NAM_JOB_EXCHANGE,
		eventbus.ExchangeTypeTopic,
		eventbus_constant.ONE_DAY_JOB_ROUTING_KEY,
		eventbus_constant.REPEATED_JOB_QUEUE,
		1,
	)

	go func() {
		err := consumer.Consume(ctx, func(d rabbitmq.Delivery) (action rabbitmq.Action) {
			return c.cronJobHandler.CreateTodayRepeatedWorks(ctx, d)
		})

		if err != nil {
			c.logger.Error("Failed to consume create today repeated works job messages", "", zap.Error(err))
			return
		}
	}()
}
