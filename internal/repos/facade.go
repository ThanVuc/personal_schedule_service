package repos

import (
	"context"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/internal/grpc/models"
	"personal_schedule_service/proto/personal_schedule"
	"time"

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
		GetLabelByKey(ctx context.Context, key string) (*collection.Label, error)
		CheckLabelExistence(ctx context.Context, id string) (bool, error)
	}

	GoalRepo interface {
		GetGoals(ctx context.Context, req *personal_schedule.GetGoalsRequest) ([]AggregatedGoal, int32, error)
		CreateGoal(ctx context.Context, goal *collection.Goal) (bson.ObjectID, error)
		UpdateGoal(ctx context.Context, goalID bson.ObjectID, updates bson.M) error
		GetTasksByGoalID(ctx context.Context, goalID bson.ObjectID) ([]collection.GoalTask, error)
		BulkWriteTasks(ctx context.Context, operations []mongo.WriteModel) (*mongo.BulkWriteResult, error)
		GetGoalByID(ctx context.Context, goalID bson.ObjectID) (*collection.Goal, error)
		GetAggregatedGoalByID(ctx context.Context, goalID bson.ObjectID) (*AggregatedGoal, error)
		DeleteTasksByGoalID(ctx context.Context, goalID bson.ObjectID) error
		DeleteGoal(ctx context.Context, goalID bson.ObjectID) error
		GetGoalsForDialog(ctx context.Context, userID string) ([]collection.Goal, error)
		UpdateGoalField(ctx context.Context, goalID bson.ObjectID, fieldName string, labelID bson.ObjectID) error
	}

	WorkRepo interface {
		GetWorkByID(ctx context.Context, workID bson.ObjectID) (*collection.Work, error)
		CreateWork(ctx context.Context, work *collection.Work) (bson.ObjectID, error)
		UpdateWork(ctx context.Context, workID bson.ObjectID, work *collection.Work) error
		GetSubTasksByWorkID(ctx context.Context, workID bson.ObjectID) ([]collection.SubTask, error)
		BulkWriteSubTasks(ctx context.Context, operations []mongo.WriteModel) (*mongo.BulkWriteResult, error)
		GetWorks(ctx context.Context, req *personal_schedule.GetWorksRequest) ([]AggregatedWork, error)
		GetAggregatedWorkByID(ctx context.Context, workID bson.ObjectID) (*AggregatedWork, error)
		CountOverlappingWorks(ctx context.Context, userID string, startDate, endDate int64, excludeWorkID *bson.ObjectID) (int64, error)
		DeleteSubTaskByWorkID(ctx context.Context, workID bson.ObjectID) error
		DeleteWork(ctx context.Context, workID bson.ObjectID) error
		DeleteDraftsByDate(ctx context.Context, userID string, startDate, endDate time.Time) error
		GetAggregatedWorksByDateRangeMs(ctx context.Context, userID string, startMs, endMs int64) ([]AggregatedWork, error)
		GetWorksByDateRangeMs(ctx context.Context, userID string, startMs, endMs int64) ([]collection.Work, error)
		BulkInsertWorks(ctx context.Context, works []interface{}) error
		BulkInsertSubTasks(ctx context.Context, subTasks []interface{}) error
		GetLabelsByTypeIDs(ctx context.Context, typeID int32) ([]collection.Label, error)
		UpdateWorkField(ctx context.Context, workID bson.ObjectID, fieldName string, labelID bson.ObjectID) error
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

func NewWorkRepo() WorkRepo {
	return &workRepo{
		logger:         global.Logger,
		mongoConnector: global.MongoDbConntector,
	}
}
