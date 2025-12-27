package initialize

import (
	"context"
	"personal_schedule_service/global"
	"sync"

	"github.com/thanvuc/go-core-lib/log"
	"go.uber.org/zap"
)

func initConfigAndResources() error {
	loadConfig()
	initLogger()
	initMongoDB()
	initRedis()
	initEventBus()
	initCronjobManager()

	return nil
}

func startGrpcSerivces(ctx context.Context, wg *sync.WaitGroup) {
	personalScheduleService := NewPersonalScheduleService()
	personalScheduleService.runServers(ctx, wg)
}

func gracefulShutdown(wg *sync.WaitGroup, logger log.Logger) {
	wg.Add(1)
	global.CronJobManager.Shutdown(wg)

	wg.Add(1)
	err := global.MongoDbConntector.GracefulClose(context.Background(), wg)
	handleError(logger, err, "MongoDB connection closed successfully")

	wg.Add(1)
	err = global.RedisDb.Close(wg)
	handleError(logger, err, "Redis connection closed successfully")

	wg.Add(1)
	global.EventBusConnector.Close(wg)

	wg.Add(1)
	err = logger.Sync(wg)
	handleError(logger, err, "Logger synced successfully")

	wg.Wait()
}

func handleError(logger log.Logger, err error, successMessage string) {
	if err != nil {
		logger.Error("An error occurred", "", zap.Error(err))
	} else {
		logger.Info(successMessage, "")
	}
}
