package validation

import (
	"context"
	"fmt"
	"personal_schedule_service/internal/collection"
	labels_constant "personal_schedule_service/internal/constant/labels"
	event_models "personal_schedule_service/internal/eventbus/models"
	"personal_schedule_service/internal/grpc/utils"
	"personal_schedule_service/internal/repos"
	app_error "personal_schedule_service/pkg/settings/error"
	"personal_schedule_service/proto/common"
	"personal_schedule_service/proto/personal_schedule"
	"time"

	"github.com/thanvuc/go-core-lib/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

type workValidator struct {
	workRepo  repos.WorkRepo
	labelRepo repos.LabelRepo
	logger    log.Logger
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
			return NewValidationError(common.ErrorCode_ERROR_CODE_INTERNAL_ERROR, app_error.EndDateBeforeStart, "EndDate must be after StartDate")
		}

		if *req.StartDate == req.EndDate {
			return NewValidationError(common.ErrorCode_ERROR_CODE_INTERNAL_ERROR, app_error.ZeroDuration, "work duration cannot be zero")
		}

		repeatLabel, _ := wv.workRepo.GetLabelByKey(ctx, labels_constant.LabelRepeated)

		var isRepeated bool
		if repeatLabel != nil {
			isRepeated = req.TypeId == repeatLabel.ID.Hex()
		} else {
			isRepeated = false
		}

		if isRepeated {
			if req.RepeatStartDate == nil || req.RepeatEndDate == nil {
				return NewValidationError(common.ErrorCode_ERROR_CODE_INTERNAL_ERROR, app_error.RepeatedWorkMissingDates, "non-repeated work cannot have repeat dates")
			}
			if *req.RepeatStartDate >= *req.RepeatEndDate {
				return NewValidationError(common.ErrorCode_ERROR_CODE_INTERNAL_ERROR, app_error.RepeatedWorkInvalidDates, "invalid repeat dates for non-repeated work")
			}
			baseStart := time.UnixMilli(*req.StartDate).UTC()
			baseEnd := time.UnixMilli(req.EndDate).UTC()
			duration := baseEnd.Sub(baseStart)

			hour, min, sec := baseStart.Clock()

			loopDate := time.UnixMilli(*req.RepeatStartDate).UTC()
			limitDate := time.UnixMilli(*req.RepeatEndDate).UTC()

			firstInstanceStart := time.Date(
				loopDate.Year(), loopDate.Month(), loopDate.Day(),
				hour, min, sec, 0, time.UTC,
			)

			lastInstanceStart := time.Date(
				limitDate.Year(), limitDate.Month(), limitDate.Day(),
				hour, min, sec, 0, time.UTC,
			)

			rangeStart := firstInstanceStart
			rangeEnd := lastInstanceStart.Add(duration)

			var excludeWorkID *bson.ObjectID
			if req.Id != nil && *req.Id != "" {
				workID, err := bson.ObjectIDFromHex(*req.Id)
				if err == nil {
					excludeWorkID = &workID
				}
			}

			existingWorks, err := wv.workRepo.GetWorksInRange(
				ctx,
				req.UserId,
				rangeStart.UnixMilli(),
				rangeEnd.UnixMilli(),
				excludeWorkID,
			)

			if err != nil {
				return NewValidationError(
					common.ErrorCode_ERROR_CODE_DATABASE_ERROR,
					app_error.TimeOverlap,
					"error retrieving works for overlap check",
				)
			}

			for !utils.TruncateToDay(loopDate).After(utils.TruncateToDay(limitDate)) {

				y, m, d := loopDate.Date()

				instanceStart := time.Date(y, m, d, hour, min, sec, 0, time.UTC)
				instanceEnd := instanceStart.Add(duration)

				for _, w := range existingWorks {

					if w.StartDate == nil {
						continue
					}

					existingStart := *w.StartDate
					existingEnd := w.EndDate

					if instanceStart.Before(existingEnd) &&
						instanceEnd.After(existingStart) {

						return NewValidationError(
							common.ErrorCode_ERROR_CODE_DATABASE_ERROR,
							app_error.TimeOverlap,
							fmt.Sprintf(
								"work overlaps on %s",
								instanceStart.Format("2006-01-02"),
							),
						)
					}
				}

				loopDate = loopDate.AddDate(0, 0, 1)
			}
		}

		if !isRepeated {
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
				return NewValidationError(common.ErrorCode_ERROR_CODE_DATABASE_ERROR, app_error.TimeOverlap, "error checking overlapping works")
			}
			if count > 0 {
				return NewValidationError(common.ErrorCode_ERROR_CODE_DATABASE_ERROR, app_error.TimeOverlap, "work time overlaps with existing work")
			}
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

func (wv *workValidator) ValidatePrompts(req *personal_schedule.GenerateWorksByAIRequest) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	errors := make([]string, 0)
	// validate prompts
	for i, prompt := range req.Prompts {
		if len(prompt) <= 20 {
			errors = append(errors, fmt.Sprintf("prompt at index %d is too short, must be at least 20 characters", i))
		}
		if len(prompt) > 500 {
			errors = append(errors, fmt.Sprintf("prompt at index %d is too long, must be at most 500 characters", i))
		}
	}

	// validate local date
	_, err := time.Parse("2006-01-02", req.LocalDate)
	if err != nil {
		return fmt.Errorf("local_date must be in format yyyy-mm-dd")
	}

	// validate additional context if provided
	if req.AdditionalContext != "" {
		if len(req.AdditionalContext) < 10 {
			errors = append(errors, "additional_context is too short, must be at least 10 characters")
		}
		if len(req.AdditionalContext) > 1000 {
			errors = append(errors, "additional_context is too long, must be at most 1000 characters")
		}
	}

	if len(errors) > 0 {
		wv.logger.Error(
			"Validation errors in GenerateWorksByAIRequest",
			"",
			zap.Error(fmt.Errorf("%v", errors)),
		)
		return fmt.Errorf("validation errors")
	}

	return nil
}

func (vw *workValidator) ValidateWorkMessages(ctx context.Context, labelMap map[string]collection.Label, workMessages []event_models.WorkMessage) error {
	for _, workMessage := range workMessages {
		if _, exists := labelMap[workMessage.DifficultyKey]; !exists {
			return fmt.Errorf("invalid difficulty key: %s", workMessage.DifficultyKey)
		}

		if _, exists := labelMap[workMessage.PriorityKey]; !exists {
			return fmt.Errorf("invalid priority key: %s", workMessage.PriorityKey)
		}

		if _, exists := labelMap[workMessage.CategoryKey]; !exists {
			return fmt.Errorf("invalid category key: %s", workMessage.CategoryKey)
		}

		if workMessage.StartDate >= workMessage.EndDate {
			return fmt.Errorf("end date must be after start date for work: %s", workMessage.Name)
		}
	}

	return nil
}
