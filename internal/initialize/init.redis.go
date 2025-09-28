package initialize

import (
	"fmt"
	"personal_schedule_service/global"
	"strconv"

	"github.com/thanvuc/go-core-lib/cache"
)

/*
@Author: Sinh
@Date: 2025/6/1
@Description: Initialize the Redis database connection.
This function uses the go-redis package to create a Redis client.
It constructs the connection string using the configuration from global.Config.Redis.
@Note: The Redis client is stored in global.RedisDb for use throughout the application.
*/

func initRedis() {
	redisConfig := global.Config.Redis
	println("HOST: " + fmt.Sprintf("%s:%s", redisConfig.Host, strconv.Itoa(redisConfig.Port)))
	redisClient := cache.NewRedisCache(cache.Config{
		Addr:     fmt.Sprintf("%s:%s", redisConfig.Host, strconv.Itoa(redisConfig.Port)),
		DB:       redisConfig.DB,
		Password: redisConfig.Password,
		PoolSize: redisConfig.PoolSize,
		MinIdle:  redisConfig.MinIdle,
	})

	if err := redisClient.Ping(); err != nil {
		global.Logger.Error("Failed to connect to Redis", "")
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	} else {
		global.Logger.Info("Redis connection established successfully", "")
	}

	global.RedisDb = redisClient
}
