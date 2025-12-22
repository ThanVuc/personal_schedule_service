package consumer

import (
	"context"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/eventbus/publisher"
	"personal_schedule_service/internal/grpc/wire"
)

func RunConsumer(ctx context.Context) {
	dlqPublisher := publisher.NewDLQPublisher()

	syncAuthDBConsumer := &SyncAuthDBConsumer{
		logger:       global.Logger,
		dlqPublisher: dlqPublisher,
		handler:      wire.InjectSyncAuthHandler(),
	}

	cronJobConsumer := NewCronJobConsumer(global.Logger)

	syncAuthDBConsumer.ConsumeUserDB(ctx)
	cronJobConsumer.ConsumeGiveUpJob(ctx)
	cronJobConsumer.ConsumeCreateTodayRepeatedWorksJob(ctx)

	global.Logger.Info("Sync Auth DB Consumer started", "")
}
