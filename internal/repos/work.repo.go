package repos

import (
	"context"
	"fmt"
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/proto/personal_schedule"

	"github.com/thanvuc/go-core-lib/log"
	"github.com/thanvuc/go-core-lib/mongolib"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

type workRepo struct {
	logger         log.Logger
	mongoConnector *mongolib.MongoConnector
}

type AggregatedWork struct {
	ID                  bson.ObjectID      `bson:"_id"`
	Name                string             `bson:"name"`
	ShortDescriptions   *string            `bson:"short_descriptions,omitempty"`
	DetailedDescription *string            `bson:"detailed_description,omitempty"`
	StartDate           *int64             `bson:"start_date,omitempty"`
	EndDate             int64              `bson:"end_date"`
	UserID              string             `bson:"user_id"`
	Status              []collection.Label `bson:"statusInfo"`
	Priority            []collection.Label `bson:"priorityInfo"`
	Difficulty          []collection.Label `bson:"difficultyInfo"`
	Type                []collection.Label `bson:"typeInfo"`
	Category            []collection.Label `bson:"categoryInfo"`
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

func (wr *workRepo) GetWorks(ctx context.Context, req *personal_schedule.GetWorksRequest) ([]AggregatedWork, error) {
	coll := wr.mongoConnector.GetCollection(collection.WorksCollection)

	dateFilter := bson.M{}
	if req.FromDate != nil {
		dateFilter["$gte"] = *req.FromDate
	}
	if req.ToDate != nil {
		dateFilter["$lte"] = *req.ToDate
	}

	fmt.Println("startDate:", &req.FromDate)
	fmt.Println("endDate:", &req.ToDate)

	matchFilter := bson.D{
		{Key: "user_id", Value: req.UserId},
		{Key: "end_date", Value: dateFilter},
	}

	if req.Search != nil && *req.Search != "" {
		matchFilter = append(matchFilter, bson.E{
			Key: "$text", Value: bson.M{
				"$search": *req.Search,
			},
		})
	}
	if req.StatusId != nil && *req.StatusId != "" {
		if objID, err := bson.ObjectIDFromHex(*req.StatusId); err == nil {
			matchFilter = append(matchFilter, bson.E{Key: "status_id", Value: objID})
		}
	}
	if req.DifficultyId != nil && *req.DifficultyId != "" {
		if objID, err := bson.ObjectIDFromHex(*req.DifficultyId); err == nil {
			matchFilter = append(matchFilter, bson.E{Key: "difficulty_id", Value: objID})
		}
	}
	if req.PriorityId != nil && *req.PriorityId != "" {
		if objID, err := bson.ObjectIDFromHex(*req.PriorityId); err == nil {
			matchFilter = append(matchFilter, bson.E{Key: "priority_id", Value: objID})
		}
	}
	if req.TypeId != nil && *req.TypeId != "" {
		if objID, err := bson.ObjectIDFromHex(*req.TypeId); err == nil {
			matchFilter = append(matchFilter, bson.E{Key: "type_id", Value: objID})
		}
	}
	if req.CategoryId != nil && *req.CategoryId != "" {
		if objID, err := bson.ObjectIDFromHex(*req.CategoryId); err == nil {
			matchFilter = append(matchFilter, bson.E{Key: "category_id", Value: objID})
		}
	}

	lookupStatus := bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         collection.LabelsCollection,
			"localField":   "status_id",
			"foreignField": "_id",
			"as":           "statusInfo",
		},
	}}
	lookupPriority := bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         collection.LabelsCollection,
			"localField":   "priority_id",
			"foreignField": "_id",
			"as":           "priorityInfo",
		},
	}}
	lookupDifficulty := bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         collection.LabelsCollection,
			"localField":   "difficulty_id",
			"foreignField": "_id",
			"as":           "difficultyInfo",
		},
	}}
	lookupCategory := bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         collection.LabelsCollection,
			"localField":   "category_id",
			"foreignField": "_id",
			"as":           "categoryInfo",
		},
	}}
	lookupType := bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         collection.LabelsCollection,
			"localField":   "type_id",
			"foreignField": "_id",
			"as":           "typeInfo",
		},
	}}

	sortStage := bson.D{
		{Key: "$sort", Value: bson.M{"end_date": 1}},
	}

	matchStage := bson.D{
		{Key: "$match", Value: matchFilter},
	}

	pipeline := mongo.Pipeline{
		matchStage,
		lookupStatus,
		lookupDifficulty,
		lookupPriority,
		lookupCategory,
		lookupType,
		sortStage,
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var works []AggregatedWork
	if err = cursor.All(ctx, &works); err != nil {
		return nil, err
	}

	return works, nil
}

func (wr *workRepo) CountOverlappingWorks(ctx context.Context, userID string, startDate, endDate int64, excludeWorkID *bson.ObjectID) (int64, error) {
	coll := wr.mongoConnector.GetCollection(collection.WorksCollection)

	filter := bson.D{
		{Key: "user_id", Value: userID},
		{Key: "start_date", Value: bson.M{"$lt": endDate}},
		{Key: "end_date", Value: bson.M{"$gt": startDate}},
	}

	if excludeWorkID != nil {
		filter = append(filter, bson.E{
			Key: "_id", Value: bson.M{"$ne": excludeWorkID},
		})
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r workRepo) GetAggregatedWorkByID(ctx context.Context, workID bson.ObjectID) (*AggregatedWork, error) {
	workCollection := r.mongoConnector.GetCollection(collection.WorksCollection)
	matchStage := bson.M{"_id": workID}
	lookupStatus := bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         collection.LabelsCollection,
			"localField":   "status_id",
			"foreignField": "_id",
			"as":           "statusInfo",
		},
	}}
	lookupPriority := bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         collection.LabelsCollection,
			"localField":   "priority_id",
			"foreignField": "_id",
			"as":           "priorityInfo",
		},
	}}
	lookupDifficulty := bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         collection.LabelsCollection,
			"localField":   "difficulty_id",
			"foreignField": "_id",
			"as":           "difficultyInfo",
		},
	}}
	lookupCategory := bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         collection.LabelsCollection,
			"localField":   "category_id",
			"foreignField": "_id",
			"as":           "categoryInfo",
		},
	}}
	lookupType := bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         collection.LabelsCollection,
			"localField":   "type_id",
			"foreignField": "_id",
			"as":           "typeInfo",
		},
	}}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		lookupStatus,
		lookupPriority,
		lookupDifficulty,
		lookupCategory,
		lookupType,
		{{Key: "$limit", Value: 1}},
	}

	cursor, err := workCollection.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error("Failed to aggregate single work", "err", zap.Error(err), zap.String("work_id", workID.Hex()))
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []AggregatedWork
	if err = cursor.All(ctx, &result); err != nil {
		r.logger.Error("Failed to decode single work", "err", zap.Error(err))
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	return &result[0], nil
}
