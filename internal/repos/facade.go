package repos

import (
	"context"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/internal/grpc/models"
	"personal_schedule_service/proto/personal_schedule"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type (
	UserRepo interface {
		UpsertSyncUser(ctx context.Context, payload models.UserOutboxPayload, requestId string) error
	}

	LabelRepo interface {
		CountLabels(ctx context.Context) (int, error)
		InsertLabels(ctx context.Context, labels *[]collection.Label) error
		GetLabels(ctx context.Context) ([]collection.Label, error)
		GetLabelsByTypeIDs(ctx context.Context, typeIDs int32) ([]collection.Label, error)
	}

	GoalRepo interface {
		GetGoals(ctx context.Context, req *personal_schedule.GetGoalsRequest) ([]AggregatedGoal, int32, error)
		CreateGoal(ctx context.Context, goal *collection.Goal) (bson.ObjectID, error)
		UpdateGoal(ctx context.Context, goalID bson.ObjectID, updates bson.M) error
		GetTasksByGoalID(ctx context.Context, goalID bson.ObjectID) ([]collection.GoalTask, error)
		BulkWriteTasks(ctx context.Context, operations []mongo.WriteModel) (*mongo.BulkWriteResult, error)
		GetGoalByID(ctx context.Context, goalID bson.ObjectID) (*collection.Goal, error)
	}
)

func NewUserRepo() UserRepo {
	return &userRepo{
		logger:    global.Logger,
		connector: global.MongoDbConntector,
	}
}

func NewLabelRepo() LabelRepo {
	return &labelRepo{
		logger:         global.Logger,
		mongoConnector: global.MongoDbConntector,
	}
}

func NewGoalRepo() GoalRepo {
	return &goalRepo{
		logger:         global.Logger,
		mongoConnector: global.MongoDbConntector,
	}
}
