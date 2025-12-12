package validation

import (
	"context"
	"fmt"
	"personal_schedule_service/internal/repos"
	app_error "personal_schedule_service/pkg/settings/error"
	"personal_schedule_service/proto/common"
	"personal_schedule_service/proto/personal_schedule"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type workValidator struct {
	workRepo  repos.WorkRepo
	labelRepo repos.LabelRepo
}

func (wv *workValidator) checkLabel(ctx context.Context, id string, name string) error {
	if _, err := bson.ObjectIDFromHex(id); err != nil {
		return NewValidationError(common.ErrorCode_ERROR_CODE_RUN_TIME_ERROR, app_error.LabelNotFoundCode, fmt.Sprintf("invalid %s format", name))
	}
	exists, err := wv.labelRepo.CheckLabelExistence(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return NewValidationError(common.ErrorCode_ERROR_CODE_NOT_FOUND, app_error.LabelNotFoundCode, fmt.Sprintf("%s %s not found", name, id))
	}
	return nil
}

func (wv *workValidator) ValidateUpsertWork(ctx context.Context, req *personal_schedule.UpsertWorkRequest) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	if err := wv.checkLabel(ctx, req.TypeId, "TypeId"); err != nil {
		return err
	}
	if err := wv.checkLabel(ctx, req.StatusId, "StatusId"); err != nil {
		return err
	}
	if err := wv.checkLabel(ctx, req.DifficultyId, "DifficultyId"); err != nil {
		return err
	}
	if err := wv.checkLabel(ctx, req.PriorityId, "PriorityId"); err != nil {
		return err
	}
	if err := wv.checkLabel(ctx, req.CategoryId, "CategoryId"); err != nil {
		return err
	}
	if req.GoalId != nil && *req.GoalId != "" {
		if _, err := bson.ObjectIDFromHex(*req.GoalId); err != nil {
			return NewValidationError(common.ErrorCode_ERROR_CODE_NOT_FOUND, app_error.GoalNotFoundCode, "invalid GoalId")
		}
	} else {
		req.GoalId = nil
	}

	for _, task := range req.SubTasks {
		if task.Id != nil && *task.Id != "" {
			if _, err := bson.ObjectIDFromHex(*task.Id); err != nil {
				return NewValidationError(common.ErrorCode_ERROR_CODE_NOT_FOUND, app_error.SubTaskNotFound, "invalid SubTask Id")
			}
		}
	}

	if req.StartDate != nil {
		if req.EndDate <= *req.StartDate {
			return NewValidationError(common.ErrorCode_ERROR_CODE_INTERNAL_ERROR, app_error.WorkEndDateBeforeStart, "EndDate must be after StartDate")
		}

		var excludeWorkID *bson.ObjectID
		if req.Id != nil && *req.Id != "" {
			workID, err := bson.ObjectIDFromHex(*req.Id)
			if err != nil {
				return NewValidationError(common.ErrorCode_ERROR_CODE_NOT_FOUND, app_error.WorkNotFound, "invalid Work Id")
			}
			excludeWorkID = &workID
		}
		count, err := wv.workRepo.CountOverlappingWorks(ctx, req.UserId, *req.StartDate, req.EndDate, excludeWorkID)
		if err != nil {
			return NewValidationError(common.ErrorCode_ERROR_CODE_DATABASE_ERROR, app_error.WorkTimeOverlap, "error checking overlapping works")
		}
		if count > 0 {
			return NewValidationError(common.ErrorCode_ERROR_CODE_DATABASE_ERROR, app_error.WorkTimeOverlap, "work time overlaps with existing work")
		}
	}

	if req.Id != nil && *req.Id != "" {
		workID, err := bson.ObjectIDFromHex(*req.Id)
		if err != nil {
			return NewValidationError(common.ErrorCode_ERROR_CODE_NOT_FOUND, app_error.WorkNotFound, "invalid Work Id")
		}
		existingWork, err := wv.workRepo.GetWorkByID(ctx, workID)
		if err != nil {
			return NewValidationError(common.ErrorCode_ERROR_CODE_DATABASE_ERROR, app_error.WorkNotFound, "error retrieving work")
		}
		if existingWork == nil {
			return NewValidationError(common.ErrorCode_ERROR_CODE_NOT_FOUND, app_error.WorkNotFound, "work not found")
		}
		if existingWork.UserID != req.UserId {
			return NewValidationError(common.ErrorCode_ERROR_CODE_PERMISSION_DENIED, app_error.WorkForbidden, "user does not have permission to modify this work")
		}
	}

	return nil
}
