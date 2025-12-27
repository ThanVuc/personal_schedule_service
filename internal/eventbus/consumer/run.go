package consumer

import (
	"context"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/eventbus/publisher"
	"personal_schedule_service/internal/wire"
)

func RunConsumer(ctx context.Context) {
	dlqPublisher := publisher.NewDLQPublisher()

	syncAuthDBConsumer := &SyncAuthDBConsumer{
		logger:       global.Logger,
		dlqPublisher: dlqPublisher,
		handler:      wire.InjectSyncAuthHandler(),
	}

	syncAuthDBConsumer.ConsumeUserDB(ctx)

	global.Logger.Info("Sync Auth DB Consumer started", "")
}
