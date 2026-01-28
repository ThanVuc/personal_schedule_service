package initialize

import (
	"context"
	"os"
	"os/signal"
	"personal_schedule_service/global"
	cronjob_run "personal_schedule_service/internal/cronjob"
	"personal_schedule_service/internal/eventbus/consumer"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

func Run() {
	if err := InitConfigAndResources(); err != nil {
		global.Logger.Error("Failed to initialize configs and resources", "", zap.Error(err))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	// Start gRPC services
	startGrpcSerivces(ctx, wg)

	// Start event consumers
	consumer.RunConsumer(ctx)

	// start cron jobs
	go cronjob_run.RunCronJob(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	global.Logger.Info("Shutdown signal received, shutting down...", "")

	cancel()

	gracefulShutdown(wg, global.Logger)
}
