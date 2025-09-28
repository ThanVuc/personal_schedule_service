package collection

import (
	"personal_schedule_service/global"

	"go.uber.org/zap"
)

func CreateCollections() error {
	err := make([]error, 0)
	logger := global.Logger

	err = append(err, createUserCollection())
	err = append(err, createGoalCollection())
	err = append(err, createGoalTaskCollection())
	err = append(err, createLabelCollection())
	err = append(err, createWorkCollection())
	err = append(err, createSubTaskCollection())

	for _, e := range err {
		if e != nil {
			logger.Error("Error creating collection: ", "", zap.Error(e))
			return e
		}
	}
	logger.Info("All collections created or already exist", "")
	return nil
}
