package repos

import (
	"context"
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/internal/grpc/utils"
	"personal_schedule_service/proto/personal_schedule"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"

	"github.com/thanvuc/go-core-lib/log"
	"github.com/thanvuc/go-core-lib/mongolib"
)

type goalRepo struct {
	logger         log.Logger
	mongoConnector *mongolib.MongoConnector
}

type AggregatedGoal struct {
	ID                  bson.ObjectID      `bson:"_id"`
	Name                string             `bson:"name"`
	ShortDescriptions   *string            `bson:"short_descriptions,omitempty"`
	DetailedDescription *string            `bson:"detailed_description,omitempty"`
	StartDate           *time.Time         `bson:"start_date,omitempty"`
	EndDate             *time.Time         `bson:"end_date,omitempty"`
	UserID              string             `bson:"user_id"`
	Status              []collection.Label `bson:"statusInfo"`
	Priority            []collection.Label `bson:"priorityInfo"`
	Difficulty          []collection.Label `bson:"difficultyInfo"`
	CreatedAt           time.Time          `bson:"created_at"`
}

type totalCountResult struct {
	Total int32 `bson:"total"`
}

func (r *goalRepo) GetGoals(ctx context.Context, req *personal_schedule.GetGoalsRequest) ([]AggregatedGoal, int32, error) {
	goalCollection := r.mongoConnector.GetCollection(collection.GoalsCollection)
	pagination := utils.ToPagination(req.PageQuery)

	// Match conditions
	matchStage := bson.D{{Key: "user_id", Value: req.UserId}}
	if req.Search != "" {
		matchStage = append(matchStage, bson.E{
			Key: "name",
			Value: bson.M{
				"$regex": bson.Regex{Pattern: req.Search, Options: "i"},
			},
		})
	}

	if req.StatusId != "" {
		objID, err := bson.ObjectIDFromHex(req.StatusId)
		if err == nil {
			matchStage = append(matchStage, bson.E{Key: "status_id", Value: objID})
		} else {
			r.logger.Warn("Invalid filter_by_status_id format", "", zap.String("status_id", req.StatusId))
		}
	}

	// Lookup stages
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

	// Pipeline for data
	pipelineData := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		lookupStatus,
		lookupPriority,
		lookupDifficulty,
		{{Key: "$sort", Value: bson.M{"created_at": -1}}},
		{{Key: "$skip", Value: pagination.Offset}},
		{{Key: "$limit", Value: pagination.Limit}},
	}

	// Pipeline for count
	pipelineCount := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$count", Value: "total"}},
	}

	// Execute main query
	cursor, err := goalCollection.Aggregate(ctx, pipelineData)
	if err != nil {
		r.logger.Error("Failed to aggregate goals", "err", zap.Error(err))
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var goals []AggregatedGoal
	if err = cursor.All(ctx, &goals); err != nil {
		r.logger.Error("Failed to decode goals", "err", zap.Error(err))
		return nil, 0, err
	}

	// Execute count query
	countCursor, err := goalCollection.Aggregate(ctx, pipelineCount)
	totalGoals := int32(0)
	if err != nil {
		r.logger.Error("Failed to aggregate goal count", "err", zap.Error(err))
	} else {
		defer countCursor.Close(ctx)
		var countResult []totalCountResult
		if err = countCursor.All(ctx, &countResult); err == nil && len(countResult) > 0 {
			totalGoals = countResult[0].Total
		} else if err != nil {
			r.logger.Error("Failed to decode GetGoals count results", "err", zap.Error(err))
		}
	}

	return goals, totalGoals, nil
}
