package collection

import (
	"context"
	"personal_schedule_service/global"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Notification struct {
	ID               bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Mode             string        `bson:"mode" json:"mode"`
	NotifyAt         time.Time     `bson:"notify_at" json:"notify_at"`
	ShortDescription *string       `bson:"short_description,omitempty" json:"short_description,omitempty"`
	CreatedAt        time.Time     `bson:"created_at" json:"created_at"`
	LastModifiedAt   time.Time     `bson:"last_modified_at" json:"last_modified_at"`
}

type Work struct {
	ID                  bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	Name                string         `bson:"name" json:"name"`
	ShortDescriptions   *string        `bson:"short_descriptions,omitempty" json:"short_descriptions,omitempty"`
	DetailedDescription *string        `bson:"detailed_description,omitempty" json:"detailed_description,omitempty"`
	StartDate           *time.Time     `bson:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate             time.Time      `bson:"end_date" json:"end_date"`
	Notifications       []Notification `bson:"notifications,omitempty" json:"notifications,omitempty"`
	StatusID            bson.ObjectID  `bson:"status_id" json:"status_id"`
	DifficultyID        bson.ObjectID  `bson:"difficulty_id" json:"difficulty_id"`
	PriorityID          bson.ObjectID  `bson:"priority_id" json:"priority_id"`
	TypeID              bson.ObjectID  `bson:"type_id" json:"type_id"`
	CategoryID          bson.ObjectID  `bson:"category_id" json:"category_id"`
	UserID              string         `bson:"user_id" json:"user_id"`
	GoalID              bson.ObjectID  `bson:"goal_id" json:"goal_id"`
	CreatedAt           time.Time      `bson:"created_at" json:"created_at"`
	LastModifiedAt      time.Time      `bson:"last_modified_at" json:"last_modified_at"`
}

func (w *Work) CollectionName() string {
	return WorksCollection
}

func createWorkCollection() error {
	connector := global.MongoDbConntector
	ctx := context.Background()

	workValidator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"end_date", "status_id", "difficulty_id", "priority_id", "type_id", "category_id", "user_id", "goal_id", "created_at", "last_modified_at"},
			"properties": bson.M{
				"_id": bson.M{
					"bsonType":    "objectId",
					"description": "Work ID, primary key",
				},
				"name": bson.M{
					"bsonType":    "string",
					"description": "Work name",
				},
				"short_descriptions": bson.M{
					"bsonType":    []string{"string", "null"},
					"description": "Short description, optional",
				},
				"detailed_description": bson.M{
					"bsonType":    []string{"string", "null"},
					"description": "Detailed description, optional",
				},
				"start_date": bson.M{
					"bsonType":    []string{"date", "null"},
					"description": "Start date, optional",
				},
				"end_date": bson.M{
					"bsonType":    "date",
					"description": "End date, required",
				},
				"notifications": bson.M{
					"bsonType":    "array",
					"description": "List of notifications",
					"items": bson.M{
						"bsonType": "object",
						"required": []string{"mode", "notify_at", "created_at", "last_modified_at"},
						"properties": bson.M{
							"_id":               bson.M{"bsonType": "objectId"},
							"mode":              bson.M{"bsonType": "string"},
							"notify_at":         bson.M{"bsonType": "date"},
							"short_description": bson.M{"bsonType": []string{"string", "null"}},
							"created_at":        bson.M{"bsonType": "date"},
							"last_modified_at":  bson.M{"bsonType": "date"},
						},
					},
				},
				"status_id":        bson.M{"bsonType": "objectId"},
				"difficulty_id":    bson.M{"bsonType": "objectId"},
				"priority_id":      bson.M{"bsonType": "objectId"},
				"type_id":          bson.M{"bsonType": "objectId"},
				"category_id":      bson.M{"bsonType": "objectId"},
				"user_id":          bson.M{"bsonType": "string"},
				"goal_id":          bson.M{"bsonType": "objectId"},
				"created_at":       bson.M{"bsonType": "date"},
				"last_modified_at": bson.M{"bsonType": "date"},
			},
		},
	}

	workIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "name", Value: "text"},
				{Key: "short_descriptions", Value: "text"},
			},
			Options: options.Index().
				SetName("idx_work_text").
				SetDefaultLanguage("none").
				SetWeights(bson.D{
					{Key: "name", Value: 10},
					{Key: "short_descriptions", Value: 5},
				}),
		},
		// compound indexes
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "end_date", Value: 1},
			},
			Options: options.Index().SetName("idx_user_end_date"),
		},
		// single field indexes
		{Keys: bson.D{{Key: "status_id", Value: 1}}, Options: options.Index().SetName("idx_status")},
		{Keys: bson.D{{Key: "difficulty_id", Value: 1}}, Options: options.Index().SetName("idx_difficulty")},
		{Keys: bson.D{{Key: "priority_id", Value: 1}}, Options: options.Index().SetName("idx_priority")},
		{Keys: bson.D{{Key: "type_id", Value: 1}}, Options: options.Index().SetName("idx_type")},
		{Keys: bson.D{{Key: "category_id", Value: 1}}, Options: options.Index().SetName("idx_category")},
		{Keys: bson.D{{Key: "goal_id", Value: 1}}, Options: options.Index().SetName("idx_goal")},
	}

	return connector.CreateCollection(ctx, WorksCollection, workValidator, workIndexes)
}
