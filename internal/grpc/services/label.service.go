package services

import (
	"context"
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/internal/grpc/helper"

	"github.com/thanvuc/go-core-lib/log"
	"github.com/thanvuc/go-core-lib/mongolib"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

type labelService struct {
	mongoConnector mongolib.MongoConnector
	labelHelper    helper.LabelHelper
	logger         log.Logger
}

func (s *labelService) SeedLabels(ctx context.Context) error {
	count, err := s.mongoConnector.GetCollection(collection.LabelsCollection).CountDocuments(ctx, bson.M{})
	if err != nil {
		s.logger.Error("Cannot get labels", "", zap.Error(err))
		return err
	}

	if count > 0 {
		s.logger.Info("Labels already initialized, skip inserting labels", "")
		return nil
	}

	labels := s.labelHelper.GenerateLabel()
	collection := s.mongoConnector.GetCollection(collection.LabelsCollection)
	_, err = collection.InsertMany(ctx, labels)
	if err != nil {
		s.logger.Error("Cannot Initalize Labels", "", zap.Error(err))
		panic("Cannot Initalize Labels")
	}
	s.logger.Info("Labels initialized successfully", "")
	return nil
}
