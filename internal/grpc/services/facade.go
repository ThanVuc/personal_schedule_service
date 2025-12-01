package services

import (
	"context"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/grpc/helper"
	"personal_schedule_service/internal/grpc/mapper"
	"personal_schedule_service/internal/grpc/validation"
	"personal_schedule_service/internal/repos"
	"personal_schedule_service/proto/common"
	"personal_schedule_service/proto/personal_schedule"
)

type (
	LabelService interface {
		SeedLabels(ctx context.Context) error
		GetLabelPerTypes(ctx context.Context, req *common.EmptyRequest) (*personal_schedule.GetLabelPerTypesResponse, error)
		GetLabelsByTypeIDs(ctx context.Context, req *common.IDRequest) (*personal_schedule.GetLabelsByTypeIDsResponse, error)
		GetDefaultLabel(ctx context.Context, req *common.EmptyRequest) (*personal_schedule.GetDefaultLabelResponse, error)
	}

	GoalService interface {
		GetGoals(ctx context.Context, req *personal_schedule.GetGoalsRequest) (*personal_schedule.GetGoalsResponse, error)
		UpsertGoal(ctx context.Context, req *personal_schedule.UpsertGoalRequest) (*personal_schedule.UpsertGoalResponse, error)
		GetGoal(ctx context.Context, req *personal_schedule.GetGoalRequest) (*personal_schedule.GetGoalResponse, error)
		DeleteGoal(ctx context.Context, req *personal_schedule.DeleteGoalRequest) (*personal_schedule.DeleteGoalResponse, error)
		GetGoalsForDialog(ctx context.Context, req *personal_schedule.GetGoalsForDialogRequest) (*personal_schedule.GetGoalForDialogResponse, error)
	}

	WorkService interface {
		UpsertWork(ctx context.Context, req *personal_schedule.UpsertWorkRequest) (*personal_schedule.UpsertWorkResponse, error)
		GetWorks(ctx context.Context, req *personal_schedule.GetWorksRequest) (*personal_schedule.GetWorksResponse, error)
		GetWork(ctx context.Context, req *personal_schedule.GetWorkRequest) (*personal_schedule.GetWorkResponse, error)
		DeleteWork(ctx context.Context, req *personal_schedule.DeleteWorkRequest) (*personal_schedule.DeleteWorkResponse, error)
	}
)

func NewLabelService(
	labelRepo repos.LabelRepo,
	labelMapper mapper.LabelMapper,
) LabelService {
	return &labelService{
		labelHelper: helper.NewLabelHelper(),
		logger:      global.Logger,
		labelRepo:   labelRepo,
		labelMapper: labelMapper,
	}
}

func NewGoalService(
	goalRepo repos.GoalRepo,
	goalMapper mapper.GoalMapper,
) GoalService {
	return &goalService{
		logger:         global.Logger,
		goalRepo:       goalRepo,
		goalMapper:     goalMapper,
		mongoConnector: global.MongoDbConntector,
	}
}

func NewWorkService(
	workRepo repos.WorkRepo,
	workMapper mapper.WorkMapper,
	validator validation.WorkValidator,
) WorkService {
	return &workService{
		logger:         global.Logger,
		workRepo:       workRepo,
		workMapper:     workMapper,
		mongoConnector: global.MongoDbConntector,
		validator:      validator,
	}
}
