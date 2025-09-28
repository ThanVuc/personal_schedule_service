package collection

import (
	"context"
	"personal_schedule_service/global"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Label struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string        `bson:"name" json:"name"`
	Meaning        *string       `bson:"meaning,omitempty" json:"meaning,omitempty"`
	Note           *string       `bson:"note,omitempty" json:"note,omitempty"`
	Color          *string       `bson:"color,omitempty" json:"color,omitempty"`
	LabelType      int           `bson:"label_type" json:"label_type"` // now integer
	CreatedAt      time.Time     `bson:"created_at" json:"created_at"`
	LastModifiedAt time.Time     `bson:"last_modified_at" json:"last_modified_at"`
}

func (l *Label) CollectionName() string {
	return LabelsCollection
}

func createLabelCollection() error {
	connector := global.MongoDbConntector
	ctx := context.Background()

	labelValidator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"name", "label_type", "created_at", "last_modified_at"},
			"properties": bson.M{
				"_id": bson.M{
					"bsonType":    "objectId",
					"description": "Label ID, primary key",
				},
				"name": bson.M{
					"bsonType":    "string",
					"description": "Label name, required",
				},
				"meaning": bson.M{
					"bsonType":    []string{"string", "null"},
					"description": "Meaning of label, optional",
				},
				"note": bson.M{
					"bsonType":    []string{"string", "null"},
					"description": "Additional note, optional",
				},
				"color": bson.M{
					"bsonType":    []string{"string", "null"},
					"description": "Color of label, optional",
				},
				"label_type": bson.M{
					"bsonType":    "int",
					"description": "Label type as integer (1,2,3...), required",
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

	labelIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "label_type", Value: 1}},
			Options: options.Index().SetName("idx_label_type"),
		},
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetName("idx_name"),
		},
	}

	return connector.CreateCollection(ctx, LabelsCollection, labelValidator, labelIndexes)
}
