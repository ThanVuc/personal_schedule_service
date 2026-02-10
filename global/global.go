package global

import (
	"personal_schedule_service/pkg/settings"
	"time"

	"github.com/thanvuc/go-core-lib/cache"
	"github.com/thanvuc/go-core-lib/cronjob"
	"github.com/thanvuc/go-core-lib/eventbus"
	"github.com/thanvuc/go-core-lib/log"
	"github.com/thanvuc/go-core-lib/mongolib"
)

/*
@Author: Sinh
@Date: 2025/6/1
@Description: This package defines global variables that are used throughout the application.
*/
var (
	Config            settings.Config
	Logger            log.Logger
	RedisDb           *cache.RedisCache
	EventBusConnector *eventbus.RabbitMQConnector
	MongoDbConntector *mongolib.MongoConnector
	CronJobManager    *cronjob.CronManager
	HCMTimeLocation   *time.Location
)
