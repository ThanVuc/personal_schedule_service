package collection

import (
	"context"
	"personal_schedule_service/global"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID             string    `bson:"_id" json:"id"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	LastModifiedAt time.Time `bson:"last_modified_at" json:"last_modified_at"`
}

func (u *User) CollectionName() string {
	return UsersCollection
}

func createUserCollection() error {
	connector := global.MongoDbConntector
	ctx := context.Background()

	userValidator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"_id", "turn_on_app_notification", "turn_on_email_notification", "created_at", "last_modified_at"},
			"properties": bson.M{
				"_id": bson.M{
					"bsonType":    "string",
					"description": "Mã định danh người dùng, bắt buộc và duy nhất",
				},
				"created_at": bson.M{
					"bsonType":    "date",
					"description": "Thời điểm tạo bản ghi, bắt buộc",
				},
				"last_modified_at": bson.M{
					"bsonType":    "date",
					"description": "Thời điểm cập nhật cuối cùng, bắt buộc",
				},
			},
		},
	}

	return connector.CreateCollection(ctx, UsersCollection, userValidator, nil)
}
