package cronjob_run

import (
	"context"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/wire"
)

func RunCronJob(ctx context.Context) {
	workCronJob := wire.InjectWorkCronJob()
	workCronJob.CreateDailyWorkCronJob(ctx)
	workCronJob.DeleteDraftWorkCronJob(ctx)
	global.Logger.Info("Cron jobs started", "")
}
