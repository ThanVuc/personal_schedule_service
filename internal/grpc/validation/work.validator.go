package validation

import (
	"context"
	"fmt"
	"personal_schedule_service/internal/repos"
	"personal_schedule_service/proto/personal_schedule"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type workValidator struct {
	workRepo repos.WorkRepo
}

func (wv *workValidator) ValidateUpsertWork(ctx context.Context, req *personal_schedule.UpsertWorkRequest) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	if _, err := bson.ObjectIDFromHex(req.StatusId); err != nil {
		return fmt.Errorf("invalid StatusId")
	}
	if _, err := bson.ObjectIDFromHex(req.DifficultyId); err != nil {
		return fmt.Errorf("invalid DifficultyId")
	}
	if _, err := bson.ObjectIDFromHex(req.PriorityId); err != nil {
		return fmt.Errorf("invalid PriorityId")
	}
	if _, err := bson.ObjectIDFromHex(req.TypeId); err != nil {
		return fmt.Errorf("invalid TypeId")
	}
	if _, err := bson.ObjectIDFromHex(req.CategoryId); err != nil {
		return fmt.Errorf("invalid CategoryId")
	}
	if req.GoalId != nil && *req.GoalId != "" {
		if _, err := bson.ObjectIDFromHex(*req.GoalId); err != nil {
			return fmt.Errorf("invalid GoalId")
		}
	} else {
		req.GoalId = nil
	}
	for _, notificationID := range req.NotificationIds {
		if _, err := bson.ObjectIDFromHex(notificationID); err != nil {
			return fmt.Errorf("invalid NotificationId %s: %v", notificationID, err)
		}
	}

	for _, task := range req.SubTasks {
		if task.Id != nil && *task.Id != "" {
			if _, err := bson.ObjectIDFromHex(*task.Id); err != nil {
				return fmt.Errorf("invalid SubTask Id %s: %v", *task.Id, err)
			}
		}
	}

	if req.StartDate != nil {
		var excludeWorkID *bson.ObjectID
		if req.Id != nil && *req.Id != "" {
			workID, err := bson.ObjectIDFromHex(*req.Id)
			if err != nil {
				return fmt.Errorf("invalid Work Id %s: %v", *req.Id, err)
			}
			excludeWorkID = &workID
		}
		count, err := wv.workRepo.CountOverlappingWorks(ctx, req.UserId, *req.StartDate, req.EndDate, excludeWorkID)
		if err != nil {
			return fmt.Errorf("error checking overlapping works: %v", err)
		}
		if count > 0 {
			return fmt.Errorf("work time overlaps with existing works")
		}
	}

	if req.Id != nil && *req.Id != "" {
		workID, err := bson.ObjectIDFromHex(*req.Id)
		if err != nil {
			return fmt.Errorf("invalid Work Id %s: %v", *req.Id, err)
		}
		existingWork, err := wv.workRepo.GetWorkByID(ctx, workID)
		if err != nil {
			return fmt.Errorf("error checking existing work: %v", err)
		}
		if existingWork == nil {
			return fmt.Errorf("work not found")
		}
		if existingWork.UserID != req.UserId {
			return fmt.Errorf("forbidden: user does not own this work")
		}
	}

	return nil
}
