package repos

import (
	"context"
	"personal_schedule_service/internal/collection"

	"github.com/thanvuc/go-core-lib/log"
	"github.com/thanvuc/go-core-lib/mongolib"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type workRepo struct {
	logger         log.Logger
	mongoConnector *mongolib.MongoConnector
}

func (wr *workRepo) GetWorkByID(ctx context.Context, workID bson.ObjectID) (*collection.Work, error) {
	coll := wr.mongoConnector.GetCollection(collection.WorksCollection)
	var work collection.Work
	err := coll.FindOne(ctx, bson.M{"_id": workID}).Decode(&work)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &work, nil
}

func (wr *workRepo) CreateWork(ctx context.Context, work *collection.Work) (bson.ObjectID, error) {
	coll := wr.mongoConnector.GetCollection(collection.WorksCollection)
	work.ID = bson.NewObjectID()
	res, err := coll.InsertOne(ctx, work)
	if err != nil {
		return bson.NilObjectID, err
	}
	return res.InsertedID.(bson.ObjectID), nil
}

func (wr *workRepo) UpdateWork(ctx context.Context, workID bson.ObjectID, updates bson.M) error {
	coll := wr.mongoConnector.GetCollection(collection.WorksCollection)
	_, err := coll.UpdateOne(ctx, bson.M{"_id": workID}, bson.M{"$set": updates})
	return err
}

func (wr *workRepo) GetSubTasksByWorkID(ctx context.Context, workID bson.ObjectID) ([]collection.SubTask, error) {
	coll := wr.mongoConnector.GetCollection(collection.SubTasksCollection)
	cursor, err := coll.Find(ctx, bson.M{"work_id": workID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subTasks []collection.SubTask
	if err = cursor.All(ctx, &subTasks); err != nil {
		return nil, err
	}
	return subTasks, nil
}

func (wr *workRepo) BulkWriteSubTasks(ctx context.Context, operations []mongo.WriteModel) (*mongo.BulkWriteResult, error) {
	if len(operations) == 0 {
		return nil, nil
	}
	coll := wr.mongoConnector.GetCollection(collection.SubTasksCollection)
	return coll.BulkWrite(ctx, operations)
}
