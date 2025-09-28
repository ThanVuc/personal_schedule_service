package publisher

import (
	"context"
	"personal_schedule_service/global"

	"github.com/thanvuc/go-core-lib/eventbus"
	"github.com/thanvuc/go-core-lib/log"
	"go.uber.org/zap"
)

type DlqPublisher struct {
	publisher eventbus.Publisher
	logger    log.Logger
}

func NewDLQPublisher() *DlqPublisher {
	publisher := eventbus.NewPublisher(
		global.EventBusConnector,
		eventbus.DLQSyncDatabaseExchange,
		eventbus.ExchangeTypeTopic,
		nil,
		nil,
		false,
	)
	return &DlqPublisher{
		publisher: publisher,
		logger:    global.Logger,
	}
}

func (d *DlqPublisher) PublishSyncUserDLQMessage(ctx context.Context, requestId string, outbox []byte) error {
	err := d.publisher.Publish(
		ctx,
		requestId,
		[]string{SyncAuthDLQ_RoutingKey},
		outbox,
		nil,
	)

	if err != nil {
		d.logger.Error("Failed to publish sync user DLQ message",
			requestId,
			zap.Error(err),
		)
		return err
	}

	return nil
}
