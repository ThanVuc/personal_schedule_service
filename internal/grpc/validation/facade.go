package validation

import (
	"context"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/collection"
	event_models "personal_schedule_service/internal/eventbus/models"
	"personal_schedule_service/internal/repos"
	"personal_schedule_service/proto/personal_schedule"
)

type (
	WorkValidator interface {
		ValidateUpsertWork(ctx context.Context, req *personal_schedule.UpsertWorkRequest) error
		ValidatePrompts(req *personal_schedule.GenerateWorksByAIRequest) error
		ValidateWorkMessages(ctx context.Context, labelMap map[string]collection.Label, workMessages []event_models.WorkMessage) error
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
		logger:    global.Logger,
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
