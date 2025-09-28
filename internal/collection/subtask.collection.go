package collection

import (
	"context"
	"personal_schedule_service/global"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SubTask struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string        `bson:"name" json:"name"`
	IsCompleted    bool          `bson:"is_completed" json:"is_completed"`
	WorkID         bson.ObjectID `bson:"work_id" json:"work_id"`
	CreatedAt      time.Time     `bson:"created_at" json:"created_at"`
	LastModifiedAt time.Time     `bson:"last_modified_at" json:"last_modified_at"`
}

func (s *SubTask) CollectionName() string {
	return SubTasksCollection
}

func createSubTaskCollection() error {
	connector := global.MongoDbConntector
	ctx := context.Background()

	subTaskValidator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"name", "is_completed", "work_id", "created_at", "last_modified_at"},
			"properties": bson.M{
				"_id": bson.M{
					"bsonType":    "objectId",
					"description": "SubTask ID, primary key",
				},
				"name": bson.M{
					"bsonType":    "string",
					"description": "SubTask name, required",
				},
				"is_completed": bson.M{
					"bsonType":    "bool",
					"description": "Completion status, required",
				},
				"work_id": bson.M{
					"bsonType":    "objectId",
					"description": "Reference to parent Work, required",
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

	subTaskIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "work_id", Value: 1}},
			Options: options.Index().SetName("idx_work"),
		},
	}

	return connector.CreateCollection(ctx, SubTasksCollection, subTaskValidator, subTaskIndexes)
}
