package repos

import (
	"context"
	"fmt"
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/internal/grpc/utils"
	"personal_schedule_service/proto/personal_schedule"
	"time"

	"github.com/thanvuc/go-core-lib/log"
	"github.com/thanvuc/go-core-lib/mongolib"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type workRepo struct {
	logger         log.Logger
	mongoConnector *mongolib.MongoConnector
}

type GoalInfo struct {
	ID   bson.ObjectID `bson:"_id"`
	Name string        `bson:"name"`
}

type AggregatedWork struct {
	ID                  bson.ObjectID      `bson:"_id"`
	Name                string             `bson:"name"`
	NameNormalized      string             `bson:"name_normalized"`
	ShortDescriptions   *string            `bson:"short_descriptions,omitempty"`
	DetailedDescription *string            `bson:"detailed_description,omitempty"`
	StartDate           *time.Time         `bson:"start_date,omitempty"`
	EndDate             time.Time          `bson:"end_date"`
	UserID              string             `bson:"user_id"`
	GoalInfo            []GoalInfo         `bson:"goalInfo"`
	Status              []collection.Label `bson:"statusInfo"`
	Priority            []collection.Label `bson:"priorityInfo"`
	Difficulty          []collection.Label `bson:"difficultyInfo"`
	Type                []collection.Label `bson:"typeInfo"`
	Category            []collection.Label `bson:"categoryInfo"`
	Overdue             []collection.Label `bson:"overdue,omitempty"`
	Draft               []collection.Label `bson:"draftInfo,omitempty"`
}

type totalCountWorksResult struct {
	Total int32 `bson:"total"`
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

func (wr *workRepo) UpdateWork(ctx context.Context, workID bson.ObjectID, work *collection.Work) error {
	coll := wr.mongoConnector.GetCollection(collection.WorksCollection)
	now := time.Now().UTC()
	updates := bson.M{
		"name":                 work.Name,
		"name_normalized":      work.NameNormalized,
		"short_descriptions":   work.ShortDescriptions,
		"detailed_description": work.DetailedDescription,
		"start_date":           work.StartDate,
		"end_date":             work.EndDate,
		"status_id":            work.StatusID,
		"difficulty_id":        work.DifficultyID,
		"priority_id":          work.PriorityID,
		"type_id":              work.TypeID,
		"category_id":          work.CategoryID,
		"draft_id":             work.DraftID,
		"goal_id":              work.GoalID,
		"last_modified_at":     now,
	}
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

func (wr *workRepo) GetWorks(ctx context.Context, req *personal_schedule.GetWorksRequest) ([]AggregatedWork, int32, error) {
	coll := wr.mongoConnector.GetCollection(collection.WorksCollection)

	var fromDate, toDate time.Time
	if req.FromDate != nil {
		fromDate = time.UnixMilli(*req.FromDate)
	}
	if req.ToDate != nil {
		toDate = time.UnixMilli(*req.ToDate)
	}

	matchFilter := bson.D{
		{Key: "user_id", Value: req.UserId},
		{Key: "start_date", Value: bson.M{
			"$gte": fromDate,
			"$lte": toDate,
		}},
	}

	if req.Search != nil && *req.Search != "" {
		searchNorm := utils.RemoveAccent(*req.Search)
		matchFilter = append(matchFilter, bson.E{
			Key: "name_normalized",
			Value: bson.M{
				"$regex":   searchNorm,
				"$options": "i",
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
	lookupDraft := bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         collection.LabelsCollection,
			"localField":   "draft_id",
			"foreignField": "_id",
			"as":           "draftInfo",
		},
	}}
	lookupGoal := bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         collection.GoalsCollection,
			"localField":   "goal_id",
			"foreignField": "_id",
			"as":           "goalInfo",
			"pipeline": bson.A{
				bson.D{{Key: "$project", Value: bson.M{"name": 1}}},
			},
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
		lookupDraft,
		sortStage,
		lookupGoal,
	}

	pipelineCount := mongo.Pipeline{
		{{Key: "$match", Value: matchFilter}},
		{{Key: "$count", Value: "total"}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var works []AggregatedWork
	if err = cursor.All(ctx, &works); err != nil {
		wr.logger.Error("Failed to decode works", "", zap.Error(err))
		return nil, 0, err
	}

	countCursor, err := coll.Aggregate(ctx, pipelineCount)
	totalWorks := int32(0)
	if err != nil {
		wr.logger.Error("Failed to aggregate works count", "", zap.Error(err))
		return nil, 0, err
	} else {
		defer countCursor.Close(ctx)
		var countResult []totalCountResult
		if err = countCursor.All(ctx, &countResult); err == nil && len(countResult) > 0 {
			totalWorks = countResult[0].Total
		} else if err != nil {
			wr.logger.Error("Failed to decode GetWorks count results", "err", zap.Error(err))
		}
	}
	fmt.Println("Total works:", totalWorks)

	return works, totalWorks, nil
}

func (wr *workRepo) CountOverlappingWorks(ctx context.Context, userID string, startDate, endDate int64, excludeWorkID *bson.ObjectID) (int64, error) {
	coll := wr.mongoConnector.GetCollection(collection.WorksCollection)

	start := time.UnixMilli(startDate)
	end := time.UnixMilli(endDate)

	filter := bson.D{
		{Key: "user_id", Value: userID},

		{Key: "start_date", Value: bson.M{"$lt": end}},
		{Key: "end_date", Value: bson.M{"$gt": start}},
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

func (wr *workRepo) GetAggregatedWorkByID(ctx context.Context, workID bson.ObjectID) (*AggregatedWork, error) {
	workCollection := wr.mongoConnector.GetCollection(collection.WorksCollection)
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
	lookupDraft := bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         collection.LabelsCollection,
			"localField":   "draft_id",
			"foreignField": "_id",
			"as":           "draftInfo",
		},
	}}
	lookupGoal := bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         collection.GoalsCollection,
			"localField":   "goal_id",
			"foreignField": "_id",
			"as":           "goalInfo",
			"pipeline": bson.A{
				bson.D{{Key: "$project", Value: bson.M{"name": 1}}},
			},
		},
	}}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		lookupStatus,
		lookupPriority,
		lookupDifficulty,
		lookupCategory,
		lookupType,
		lookupGoal,
		lookupDraft,
		{{Key: "$limit", Value: 1}},
	}

	cursor, err := workCollection.Aggregate(ctx, pipeline)
	if err != nil {
		wr.logger.Error("Failed to aggregate single work", "err", zap.Error(err), zap.String("work_id", workID.Hex()))
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []AggregatedWork
	if err = cursor.All(ctx, &result); err != nil {
		wr.logger.Error("Failed to decode single work", "err", zap.Error(err))
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	return &result[0], nil
}

func (wr *workRepo) DeleteSubTaskByWorkID(ctx context.Context, workID bson.ObjectID) error {
	coll := wr.mongoConnector.GetCollection(collection.SubTasksCollection)
	_, err := coll.DeleteMany(ctx, bson.M{"work_id": workID})
	return err
}

func (wr *workRepo) DeleteWork(ctx context.Context, workID bson.ObjectID) error {
	coll := wr.mongoConnector.GetCollection(collection.WorksCollection)
	result, err := coll.DeleteOne(ctx, bson.M{"_id": workID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (wr *workRepo) DeleteDraftsByDate(ctx context.Context, userID string, startDate, endDate time.Time) error {
	coll := wr.mongoConnector.GetCollection(collection.WorksCollection)

	filter := bson.D{
		{Key: "user_id", Value: userID},
		{Key: "draft_id", Value: bson.M{
			"$exists": true,
			"$ne":     nil,
		}},
		{Key: "$and", Value: bson.A{
			bson.D{{Key: "start_date", Value: bson.M{"$lte": endDate}}},
			bson.D{{Key: "end_date", Value: bson.M{"$gte": startDate}}},
		}},
	}
	_, err := coll.DeleteMany(ctx, filter)
	return err
}

func (wr *workRepo) GetAggregatedWorksByDateRangeMs(ctx context.Context, userID string, startMs, endMs int64) ([]AggregatedWork, error) {
	coll := wr.mongoConnector.GetCollection(collection.WorksCollection)

	// convert milliseconds to time.Time before comparing with BSON date fields
	start := time.UnixMilli(startMs)
	end := time.UnixMilli(endMs)

	matchStage := bson.D{
		{Key: "user_id", Value: userID},
		{Key: "$and", Value: bson.A{
			bson.D{{Key: "$or", Value: bson.A{
				bson.D{{Key: "start_date", Value: bson.M{"$lte": end}}},
				bson.D{{Key: "start_date", Value: nil}},
			}}},
			bson.D{{Key: "end_date", Value: bson.M{"$gte": start}}},
		}},
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
	lookupGoal := bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         collection.GoalsCollection,
			"localField":   "goal_id",
			"foreignField": "_id",
			"as":           "goalInfo",
			"pipeline": bson.A{
				bson.D{{Key: "$project", Value: bson.M{"name": 1}}},
			},
		},
	}}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		lookupStatus, lookupPriority, lookupDifficulty, lookupCategory, lookupType, lookupGoal,
		{{Key: "$sort", Value: bson.M{"start_date": 1}}},
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

func (wr *workRepo) BulkInsertWorks(ctx context.Context, works []interface{}) error {
	if len(works) == 0 {
		return nil
	}
	coll := wr.mongoConnector.GetCollection(collection.WorksCollection)
	_, err := coll.InsertMany(ctx, works)
	wr.logger.Info("BulkInsertWorks", "", zap.Int("count", len(works)))
	return err
}
func (wr *workRepo) BulkInsertSubTasks(ctx context.Context, subTasks []interface{}) error {
	if len(subTasks) == 0 {
		return nil
	}
	coll := wr.mongoConnector.GetCollection(collection.SubTasksCollection)
	_, err := coll.InsertMany(ctx, subTasks)
	wr.logger.Info("BulkInsertSubTasks", "", zap.Int("count", len(subTasks)))
	return err
}

func (wr *workRepo) GetLabelsByTypeIDs(ctx context.Context, typeID int32) ([]collection.Label, error) {
	var labels []collection.Label
	collection := wr.mongoConnector.GetCollection(collection.LabelsCollection)
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

func (wr *workRepo) UpdateWorkField(ctx context.Context, workID bson.ObjectID, fieldName string, labelID bson.ObjectID) error {
	coll := wr.mongoConnector.GetCollection(collection.WorksCollection)

	update := bson.M{
		"$set": bson.M{
			fieldName:          labelID,
			"last_modified_at": time.Now(),
		},
	}

	_, err := coll.UpdateOne(ctx, bson.M{"_id": workID}, update)
	wr.logger.Info("UpdateWorkField", "", zap.Any("value", labelID))
	return err
}

func (wr *workRepo) GetLabelByKey(ctx context.Context, key string) (*collection.Label, error) {
	coll := wr.mongoConnector.GetCollection(collection.LabelsCollection)
	var label collection.Label
	err := coll.FindOne(ctx, bson.M{"key": key}).Decode(&label)
	if err != nil {
		return nil, err
	}
	return &label, nil
}

func (wr *workRepo) CommitRecoveryDrafts(ctx context.Context, req *personal_schedule.CommitRecoveryDraftsRequest, draftID bson.ObjectID) error {
	coll := wr.mongoConnector.GetCollection(collection.WorksCollection)
	filler := bson.M{
		"_id":      bson.M{"$in": req.WorkIds},
		"user_id":  req.UserId,
		"draft_id": draftID,
	}
	update := bson.M{
		"$unset": bson.M{
			"draft_id": "",
		},
	}
	_, err := coll.UpdateMany(ctx, filler, update)
	wr.logger.Info("CommitRecoveryDrafts", "", zap.Int("work_count", len(req.WorkIds)))
	return err
}
