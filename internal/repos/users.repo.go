package repos

import (
	"context"
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/internal/grpc/models"
	"time"

	"github.com/thanvuc/go-core-lib/log"
	"github.com/thanvuc/go-core-lib/mongolib"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type userRepo struct {
	logger    log.Logger
	connector *mongolib.MongoConnector
}

func (r *userRepo) UpsertSyncUser(ctx context.Context, payload models.UserOutboxPayload, requestId string) error {
	collection := r.connector.GetCollection(collection.UsersCollection)
	filter := bson.M{"_id": payload.UserID}

	update := bson.M{
		"$set": bson.M{
			"_id":                        payload.UserID,
			"email":                      payload.Email,
			"turn_on_app_notification":   true,
			"turn_on_email_notification": true,
			"created_at":                 time.Unix(payload.CreatedAt, 0),
			"last_modified_at":           time.Now(),
		},
	}

	opts := options.UpdateOne().SetUpsert(true)

	if _, err := collection.UpdateOne(ctx, filter, update, opts); err != nil {
		r.logger.Error("Failed to upsert user", requestId, zap.Error(err))
		return err
	}

	return nil
}
