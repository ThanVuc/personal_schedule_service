package consumer

import (
	"context"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/eventbus/handler"
	"personal_schedule_service/internal/eventbus/publisher"
)

func RunConsumer(ctx context.Context) {
	dlqPublisher := publisher.NewDLQPublisher()

	syncAuthDBConsumer := &SyncAuthDBConsumer{
		logger:       global.Logger,
		dlqPublisher: dlqPublisher,
		handler:      handler.NewSyncAuthHandler(),
	}

	syncAuthDBConsumer.ConsumeUserDB(ctx)
	global.Logger.Info("Sync Auth DB Consumer started", "")
}
