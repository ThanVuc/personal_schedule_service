package collection

import (
	"context"
	"personal_schedule_service/global"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type GoalTask struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string        `bson:"name" json:"name"`
	IsCompleted    bool          `bson:"is_completed" json:"is_completed"`
	GoalID         bson.ObjectID `bson:"goal_id" json:"goal_id"`
	CreatedAt      time.Time     `bson:"created_at" json:"created_at"`
	LastModifiedAt time.Time     `bson:"last_modified_at" json:"last_modified_at"`
}

func (t *GoalTask) CollectionName() string {
	return GoalTasksCollection
}

func createGoalTaskCollection() error {
	connector := global.MongoDbConntector
	ctx := context.Background()

	goalTaskValidator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"name", "is_completed", "goal_id", "created_at", "last_modified_at"},
			"properties": bson.M{
				"_id": bson.M{
					"bsonType":    "objectId",
					"description": "Task ID, primary key",
				},
				"name": bson.M{
					"bsonType":    "string",
					"description": "Task name, required",
				},
				"is_completed": bson.M{
					"bsonType":    "bool",
					"description": "Completion status, required",
				},
				"goal_id": bson.M{
					"bsonType":    "objectId",
					"description": "Reference to parent goal, required",
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

	goalTaskIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "goal_id", Value: 1}},
			Options: options.Index().SetName("idx_goal"),
		},
	}

	return connector.CreateCollection(ctx, GoalTasksCollection, goalTaskValidator, goalTaskIndexes)
}
