package repos

import (
	"context"
	"personal_schedule_service/internal/collection"

	"github.com/thanvuc/go-core-lib/log"
	"github.com/thanvuc/go-core-lib/mongolib"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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
	if labels == nil || len(*labels) == 0 {
		return nil
	}

	collection := lr.mongoConnector.GetCollection(collection.LabelsCollection)
	_, err := collection.InsertMany(ctx, *labels)
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

func (lr *labelRepo) GetLabelsByTypeIDs(ctx context.Context, typeID int32) ([]collection.Label, error) {
	var labels []collection.Label
	collection := lr.mongoConnector.GetCollection(collection.LabelsCollection)
	println("typeID:", typeID)
	filter := bson.M{"label_type": typeID}
	options := options.Find().SetSort(bson.D{{Key: "color", Value: 1}})
	cursor, err := collection.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &labels); err != nil {
		return nil, err
	}
	return labels, nil
}

func (lr *labelRepo) GetLabelByKey(ctx context.Context, key string) (*collection.Label, error) {
	coll := lr.mongoConnector.GetCollection(collection.LabelsCollection)
	var label collection.Label
	err := coll.FindOne(ctx, bson.M{"key": key}).Decode(&label)
	if err != nil {
		return nil, err
	}
	return &label, nil
}

func (lr *labelRepo) CheckLabelExistence(ctx context.Context, id string) (bool, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return false, nil
	}

	coll := lr.mongoConnector.GetCollection(collection.LabelsCollection)
	count, err := coll.CountDocuments(ctx, bson.M{"_id": oid})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (lr *labelRepo) CountGoalByLabelKey(ctx context.Context, key string) (int64, error) {
	coll := lr.mongoConnector.GetCollection(collection.GoalsCollection)
	count, err := coll.CountDocuments(ctx, bson.M{"key": key})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (lr *labelRepo) GetLabelByID(ctx context.Context, labelID bson.ObjectID) (*collection.Label, error) {
	coll := lr.mongoConnector.GetCollection(collection.LabelsCollection)
	var label collection.Label
	err := coll.FindOne(ctx, bson.M{"_id": labelID}).Decode(&label)
	if err != nil {
		return nil, err
	}
	return &label, nil
}
