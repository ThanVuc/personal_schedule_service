package collection

import (
	"context"
	"personal_schedule_service/global"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Goal struct {
	ID                  bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name                string        `bson:"name" json:"name"`
	ShortDescriptions   *string       `bson:"short_descriptions,omitempty" json:"short_descriptions,omitempty"`
	DetailedDescription *string       `bson:"detailed_description,omitempty" json:"detailed_description,omitempty"`
	StartDate           *time.Time    `bson:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate             *time.Time    `bson:"end_date,omitempty" json:"end_date,omitempty"`
	StatusID            bson.ObjectID `bson:"status_id" json:"status_id"`
	DifficultyID        bson.ObjectID `bson:"difficulty_id" json:"difficulty_id"`
	PriorityID          bson.ObjectID `bson:"priority_id" json:"priority_id"`
	UserID              string        `bson:"user_id" json:"user_id"`
	CreatedAt           time.Time     `bson:"created_at" json:"created_at"`
	LastModifiedAt      time.Time     `bson:"last_modified_at" json:"last_modified_at"`
}

func (g *Goal) CollectionName() string {
	return GoalsCollection
}

func createGoalCollection() error {
	connector := global.MongoDbConntector
	ctx := context.Background()

	goalValidator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"name", "status_id", "difficulty_id", "priority_id", "user_id", "created_at", "last_modified_at"},
			"properties": bson.M{
				"_id": bson.M{
					"bsonType":    "objectId",
					"description": "Goal ID, primary key",
				},
				"name": bson.M{
					"bsonType":    "string",
					"description": "Goal name, required",
				},
				"short_descriptions": bson.M{
					"bsonType":    []string{"string", "null"},
					"description": "Short description, can be null",
				},
				"detailed_description": bson.M{
					"bsonType":    []string{"string", "null"},
					"description": "Detailed description, can be null",
				},
				"start_date": bson.M{
					"bsonType":    []string{"date", "null"},
					"description": "Start date, can be null",
				},
				"end_date": bson.M{
					"bsonType":    []string{"date", "null"},
					"description": "End date, can be null",
				},
				"status_id": bson.M{
					"bsonType":    "objectId",
					"description": "Reference to status, required",
				},
				"difficulty_id": bson.M{
					"bsonType":    "objectId",
					"description": "Reference to difficulty, required",
				},
				"priority_id": bson.M{
					"bsonType":    "objectId",
					"description": "Reference to priority, required",
				},
				"user_id": bson.M{
					"bsonType":    "string",
					"description": "Reference to user, required",
				},
				"created_at": bson.M{
					"bsonType":    "date",
					"description": "Creation timestamp, required",
				},
				"last_modified_at": bson.M{
					"bsonType":    "date",
					"description": "Last modification timestamp, required",
				},
			},
		},
	}

	goalIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("idx_user"),
		},
		{
			Keys:    bson.D{{Key: "status_id", Value: 1}},
			Options: options.Index().SetName("idx_status"),
		},
		{
			Keys:    bson.D{{Key: "last_modified_at", Value: 1}},
			Options: options.Index().SetName("idx_last_modified"),
		},
		{
			Keys:    bson.D{{Key: "start_date", Value: 1}},
			Options: options.Index().SetName("idx_start_date"),
		},
		{
			Keys:    bson.D{{Key: "end_date", Value: 1}},
			Options: options.Index().SetName("idx_end_date"),
		},
	}

	return connector.CreateCollection(ctx, GoalsCollection, goalValidator, goalIndexes)
}
