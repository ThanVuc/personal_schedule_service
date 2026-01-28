package handler

import (
	"context"
	"personal_schedule_service/global"
	workgeneration_constant "personal_schedule_service/internal/constant/work"
	"personal_schedule_service/internal/grpc/validation"
	"personal_schedule_service/internal/repos"

	"github.com/thanvuc/go-core-lib/eventbus"
	"github.com/thanvuc/go-core-lib/log"
	"github.com/wagslane/go-rabbitmq"
)

type WorkGenerationHandler struct {
	logger            log.Logger
	workRepo          repos.WorkRepo
	workValidator     validation.WorkValidator
	eventbusConnector *eventbus.RabbitMQConnector
}

func NewWorkGenerationHandler(
	workRepo repos.WorkRepo,
	workValidator validation.WorkValidator,
) *WorkGenerationHandler {
	return &WorkGenerationHandler{
		logger:            global.Logger,
		workRepo:          workRepo,
		eventbusConnector: global.EventBusConnector,
		workValidator:     workValidator,
	}
}

func (n *WorkGenerationHandler) ConsumeWorks(ctx context.Context, d rabbitmq.Delivery) rabbitmq.Action {
	println(string(d.Body))
	publiser := eventbus.NewPublisher(
		n.eventbusConnector,
		workgeneration_constant.NOTIFICATION_GENERATE_WORK_EXCHANGE,
		eventbus.ExchangeTypeDirect,
		nil,
		nil,
		false,
	)

	err := publiser.Publish(
		ctx,
		"work_generation_handler",
		[]string{workgeneration_constant.NOTIFICATION_GENERATE_WORK_ROUTING_KEY},
		d.Body,
		nil,
	)

	if err != nil {
		n.logger.Error("Failed to publish work generated notification", "")
		return rabbitmq.NackDiscard
	}

	return rabbitmq.Ack
}
