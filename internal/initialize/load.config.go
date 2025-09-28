package initialize

import (
	"personal_schedule_service/global"

	"github.com/thanvuc/go-core-lib/config"
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
