package initialize

import (
	"context"
	"fmt"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/internal/grpc/mapper"
	"personal_schedule_service/internal/grpc/services"
	"personal_schedule_service/internal/repos"
	"sync"
	"time"

	"github.com/thanvuc/go-core-lib/mongolib"
)

var (
	onceMongo sync.Once
)

func initMongoDB() {
	onceMongo.Do(func() {
		logger := global.Logger
		cfg := createMongoConfiguration()

		const maxRetries = 10
		const retryInterval = 6 * time.Second

		var err error
		for i := 1; i <= maxRetries; i++ {
			global.MongoDbConntector, err = mongolib.NewMongoConnector(context.Background(), cfg)
			if err == nil {
				logger.Info("MongoDB connected successfully", "")
				// create collections, validators, indexes
				err := createCollections()
				if err != nil {
					logger.Error("Failed to create collections", "")
				}

				// seed initial data
				err = seedData()
				if err != nil {
					panic("Failed to seed initial data")
				}

				return
			}

			logger.Warn(fmt.Sprintf("Failed to connect to MongoDB (attempt %d/%d): %v", i, maxRetries, err), "")
			time.Sleep(retryInterval * time.Duration(i)) // Exponential backoff
		}

		logger.Error("Could not connect to MongoDB after maximum retries", "")
		panic(fmt.Sprintf("Could not connect to MongoDB after %d attempts: %v", maxRetries, err))
	})
}

func createMongoConfiguration() mongolib.MongoConnectorConfig {
	return mongolib.MongoConnectorConfig{
		URI:      global.Config.Mongo.URI,
		Database: global.Config.Mongo.Database,
		Username: global.Config.Mongo.Username,
		Password: global.Config.Mongo.Password,
	}
}

// create necessary collections, vadators, indexes
// call the create function in each model
func createCollections() error {
	return collection.CreateCollections()
}

func seedData() error {
	labelService := services.NewLabelService(
		repos.NewLabelRepo(),
		mapper.NewLabelMapper(),
	)
	err := labelService.SeedLabels(context.Background())
	if err != nil {
		return err
	}

	return nil
}
