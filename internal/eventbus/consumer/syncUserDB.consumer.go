package consumer

import (
	"context"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/eventbus/handler"
	"personal_schedule_service/internal/eventbus/publisher"
	"personal_schedule_service/proto/common"

	"github.com/thanvuc/go-core-lib/eventbus"
	"github.com/thanvuc/go-core-lib/log"
	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type SyncAuthDBConsumer struct {
	logger       log.Logger
	dlqPublisher *publisher.DlqPublisher
	handler      *handler.SyncAuthHandler
}

func (c *SyncAuthDBConsumer) ConsumeUserDB(ctx context.Context) {
	syncConsumerDLQ := eventbus.NewConsumer(
		global.EventBusConnector,
		eventbus.SyncDatabaseExchange,
		eventbus.ExchangeTypeTopic,
		"sync.auth.user",
		"sync_user_queue",
		1,
	)

	c.logger.Info("Starting to consume messages from sync user DB queue", "")
	go func() {
		err := syncConsumerDLQ.Consume(ctx, func(d rabbitmq.Delivery) (action rabbitmq.Action) {
			requestId := d.Headers["request_id"].(string)
			outbox := &common.Outbox{}
			err := proto.Unmarshal(d.Body, outbox)
			if err != nil {
				c.logger.Error("Failed to unmarshal message from sync user DB queue", "", zap.Error(err))
				c.dlqPublisher.PublishSyncUserDLQMessage(ctx, requestId, d.Body)
				return rabbitmq.NackDiscard
			}
			err = c.handler.SyncUserDB(ctx, outbox.Payload)
			if err != nil {
				c.logger.Error("Failed to sync user DB", "", zap.Error(err))
				c.dlqPublisher.PublishSyncUserDLQMessage(ctx, requestId, d.Body)
				return rabbitmq.NackDiscard
			}
			c.logger.Info("Sync user successful", requestId, zap.String("user_id", outbox.AggregateId))
			return rabbitmq.Ack
		})

		if err != nil {
			c.logger.Error("Failed to consume messages from sync DLQ", "")
			return
		}
	}()
}
