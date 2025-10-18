package repos

import (
	"context"
	"personal_schedule_service/internal/collection"

	"github.com/thanvuc/go-core-lib/log"
	"github.com/thanvuc/go-core-lib/mongolib"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type labelRepo struct {
	logger         log.Logger
	mongoConnector *mongolib.MongoConnector
}

func (lr *labelRepo) CountLabels(ctx context.Context) (int, error) {
	count, err := lr.mongoConnector.GetCollection(collection.LabelsCollection).CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (lr *labelRepo) InsertLabels(ctx context.Context, labels *[]collection.Label) error {
	collection := lr.mongoConnector.GetCollection(collection.LabelsCollection)
	_, err := collection.InsertMany(ctx, labels)
	if err != nil {
		return err
	}
	return nil
}

func (lr *labelRepo) GetLabels(ctx context.Context) ([]collection.Label, error) {
	var labels []collection.Label
	collection := lr.mongoConnector.GetCollection(collection.LabelsCollection)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &labels); err != nil {
		return nil, err
	}

	return labels, nil
}
