package initialize

import (
	"personal_schedule_service/global"
	"time"

	_ "time/tzdata"

	"github.com/thanvuc/go-core-lib/config"
	"go.uber.org/zap"
)

/*
@Author: Sinh
@Date: 2025/6/1
@Description: Load configuration from a YAML file using Viper.
The configuration file file is loaded to the global.Config variable.
*/
func loadConfig() {
	err := config.LoadConfig(&global.Config, "./")

	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
}

func loadHCMTimeLocation() {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		global.Logger.Error("Failed to load Asia/Ho_Chi_Minh time location", "", zap.Error(err))
	}
	global.HCMTimeLocation = loc
}
