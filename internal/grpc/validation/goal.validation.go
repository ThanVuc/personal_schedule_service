package validation

import (
	"context"
	"fmt"
	labels_constant "personal_schedule_service/internal/constant/labels"
	"personal_schedule_service/internal/repos"
	app_error "personal_schedule_service/pkg/settings/error"
	"personal_schedule_service/proto/common"
	"personal_schedule_service/proto/personal_schedule"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type goalValidator struct {
	goalRepo  repos.GoalRepo
	labelRepo repos.LabelRepo
}

func (gv *goalValidator) checkLabel(ctx context.Context, id string, name string) error {
	if _, err := bson.ObjectIDFromHex(id); err != nil {
		return NewValidationError(common.ErrorCode_ERROR_CODE_RUN_TIME_ERROR, app_error.LabelNotFoundCode, fmt.Sprintf("invalid %s format", name))
	}
	exists, err := gv.labelRepo.CheckLabelExistence(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return NewValidationError(common.ErrorCode_ERROR_CODE_NOT_FOUND, app_error.LabelNotFoundCode, fmt.Sprintf("%s %s not found", name, id))
	}
	return nil
}

func (gv *goalValidator) ValidationGoal(ctx context.Context, req *personal_schedule.UpsertGoalRequest) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}
	if err := gv.checkLabel(ctx, req.StatusId, "StatusId"); err != nil {
		return err
	}
	if err := gv.checkLabel(ctx, req.DifficultyId, "DifficultyId"); err != nil {
		return err
	}
	if err := gv.checkLabel(ctx, req.PriorityId, "PriorityId"); err != nil {
		return err
	}
	if err := gv.checkLabel(ctx, req.CategoryId, "CategoryId"); err != nil {
		return err
	}

	for _, task := range req.Tasks {
		if task.Id != nil && *task.Id != "" {
			if _, err := bson.ObjectIDFromHex(*task.Id); err != nil {
				return NewValidationError(common.ErrorCode_ERROR_CODE_NOT_FOUND, app_error.SubTaskNotFound, "invalid SubTask Id")
			}
		}
	}
	if req.StartDate != nil {
		if *req.EndDate <= *req.StartDate {
			return NewValidationError(common.ErrorCode_ERROR_CODE_INTERNAL_ERROR, app_error.EndDateBeforeStart, "EndDate must be after StartDate")
		}

		if *req.StartDate == *req.EndDate {
			return NewValidationError(common.ErrorCode_ERROR_CODE_INTERNAL_ERROR, app_error.ZeroDuration, "goal duration cannot be zero")
		}
	}

	statusID, err := bson.ObjectIDFromHex(req.StatusId)
	if err != nil {
		return err
	}
	label, err := gv.labelRepo.GetLabelByID(ctx, statusID)
	if err != nil {
		return err
	}
	if label.Key == "PENDING" || label.Key == "IN_PROGRESS" {
		countPending , err := gv.labelRepo.CountGoalByLabelKey(ctx, labels_constant.LabelPending)
		if err != nil {
			return err
		}
		countInProgress , err := gv.labelRepo.CountGoalByLabelKey(ctx, labels_constant.LabelInProgress)
		if err != nil {
			return err
		}
		if countPending + countInProgress >= 5 {
			return NewValidationError(common.ErrorCode_ERROR_CODE_INTERNAL_ERROR, app_error.GoalLimitExceeded, "You can only have up to 5 goals with status PENDING or IN_PROGRESS")
		}
	}

	if req.Id != nil && *req.Id != "" {
		goalID, err := bson.ObjectIDFromHex(*req.Id)
		if err != nil {
			return NewValidationError(common.ErrorCode_ERROR_CODE_NOT_FOUND, app_error.GoalNotFoundCode, "invalid goal Id")
		}
		existingGoal, err := gv.goalRepo.GetGoalByID(ctx, goalID)
		if err != nil {
			return NewValidationError(common.ErrorCode_ERROR_CODE_DATABASE_ERROR, app_error.GoalNotFoundCode, "error retrieving goal")
		}
		if existingGoal == nil {
			return NewValidationError(common.ErrorCode_ERROR_CODE_NOT_FOUND, app_error.GoalNotFoundCode, "goal not found")
		}
		if existingGoal.UserID != req.UserId {
			return NewValidationError(common.ErrorCode_ERROR_CODE_PERMISSION_DENIED, app_error.GoalForbidden, "user does not have permission to modify this goal")
		}
	}

	return nil
}
