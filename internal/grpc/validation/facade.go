package validation

import (
	"context"
	"personal_schedule_service/internal/repos"
	"personal_schedule_service/proto/personal_schedule"
)

type (
	WorkValidator interface {
		ValidateUpsertWork(ctx context.Context, req *personal_schedule.UpsertWorkRequest) error
	}
	GoalValidator interface {
		ValidationGoal(ctx context.Context, req *personal_schedule.UpsertGoalRequest) error
	}
)

func NewWorkValidator(
	workRepo repos.WorkRepo,
	label repos.LabelRepo,
) WorkValidator {
	return &workValidator{
		workRepo:  workRepo,
		labelRepo: label,
	}
}

func NewGoalValidator(
	goalRepo repos.GoalRepo,
	label repos.LabelRepo,
) GoalValidator {
	return &goalValidator{
		goalRepo:  goalRepo,
		labelRepo: label,
	}
}
