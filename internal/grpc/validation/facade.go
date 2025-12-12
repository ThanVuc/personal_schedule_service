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
