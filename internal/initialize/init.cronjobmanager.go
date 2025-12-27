package initialize

import (
	"personal_schedule_service/global"

	"github.com/thanvuc/go-core-lib/cronjob"
)

func initCronjobManager() {
	cronJobManager := cronjob.NewCronManager()
	global.CronJobManager = cronJobManager
}
