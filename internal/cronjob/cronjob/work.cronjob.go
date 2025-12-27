package cronjob

import (
	"context"
	"personal_schedule_service/global"
	cronjob_constant "personal_schedule_service/internal/cronjob/constant"
	"personal_schedule_service/internal/grpc/services"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/thanvuc/go-core-lib/cronjob"
	"github.com/thanvuc/go-core-lib/log"
	"go.uber.org/zap"
)

type WorkCronJob struct {
	cronJobManager *cronjob.CronManager
	logger         log.Logger
	workService    services.WorkService
}

func NewWorkCronJob(
	service services.WorkService,
) *WorkCronJob {
	return &WorkCronJob{
		cronJobManager: global.CronJobManager,
		logger:         global.Logger,
		workService:    service,
	}
}

func (c *WorkCronJob) CreateDailyWorkCronJob(ctx context.Context) {
	// Define the cron schedule (every day at midnight)
	jobScheduler := cronjob.NewCronScheduler(global.RedisDb, cronjob_constant.CREATE_DAILY_WORK_CRONJOB, cron.WithLocation(time.UTC))

	c.cronJobManager.AddScheduler(jobScheduler)

	// add schedule string for scheduling cronjob. eg. "0 0 * * *": every day at midnight
	// Use OTC, so minus 7 hours from UTC (get Vietnam time)
	err := jobScheduler.ScheduleCronJob("0 17 * * *", func() {
		// Logic to create daily work entries
		// call to work service to create daily work
		c.logger.Info("Executing CreateDailyWorkCronJob", "")
	})
	if err != nil {
		c.logger.Error("Failed to handle CreateDailyWorkCronJob", "", zap.Error(err))
	}
	jobScheduler.Start()
}
