// package validation

// import (
// 	"context"
// 	"fmt"
// 	"personal_schedule_service/internal/repos"
// 	"personal_schedule_service/proto/personal_schedule"

// 	"go.mongodb.org/mongo-driver/v2/bson"
// )

// type workValidator struct {
// 	workRepo repos.WorkRepo
// }

// func (wv *workValidator) ValidateUpsertWork(ctx context.Context, req *personal_schedule.UpsertWorkRequest) error {
// 	_, err := bson.ObjectIDFromHex(req.StatusId)
// 	if err != nil {
// 		return fmt.Errorf("invalid StatusId: %v", err)
// 	}
// 	_, err = bson.ObjectIDFromHex(req.DifficultyId)
// 	if err != nil {
// 		return fmt.Errorf("invalid DifficultyId: %v", err)
// 	}
// 	_, err = bson.ObjectIDFromHex(req.PriorityId)
// 	if err != nil {
// 		return fmt.Errorf("invalid PriorityId: %v", err)
// 	}
// 	_, err = bson.ObjectIDFromHex(req.TypeId)
// 	if err != nil {
// 		return fmt.Errorf("invalid TypeId: %v", err)
// 	}
// 	_, err = bson.ObjectIDFromHex(req.CategoryId)
// 	if err != nil {
// 		return fmt.Errorf("invalid CategoryId: %v", err)
// 	}
// 	_, err = bson.ObjectIDFromHex(req.GoalId)
// 	if err != nil {
// 		return fmt.Errorf("invalid GoalId: %v", err)
// 	}
// 	for _, notificationID := range req.NotificationIds {
// 		_, err = bson.ObjectIDFromHex(notificationID)
// 		if err != nil {
// 			return fmt.Errorf("invalid NotificationId %s: %v", notificationID, err)
// 		}
// 	}

// 	for _, task := range req.SubTasks {
// 		if task.Id != nil && *task.Id != "" {
// 			_, err = bson.ObjectIDFromHex(*task.Id)
// 			if err != nil {
// 				return fmt.Errorf("invalid SubTask Id %s: %v", task.Id, err)
// 			}
// 		}
// 	}

// 	if req.Id != nil && *req.Id != "" {
// 		_, err = bson.ObjectIDFromHex(*req.Id)
// 		if err != nil {
// 			return fmt.Errorf("invalid Work Id %s: %v", *req.Id, err)
// 		}

// 		workID, _ := bson.ObjectIDFromHex(*req.Id)

// 		exitingWork, err := wv.workRepo.GetWorkByID(ctx, workID)
// 		if err != nil {
// 			return fmt.Errorf("error checking existing work: %v", err)
// 		}
// 		if exitingWork == nil {
// 			return fmt.Errorf("work with Id %s does not exist", *req.Id)
// 		}
// 		if exitingWork.UserID != req.UserId {
// 			return fmt.Errorf("work with Id %s does not belong to user %s", *req.Id, req.UserId)
// 		}
// 	}
// 	return nil
// }

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


// func (wv *workValidator) ValidateUpsertWork(ctx context.Context, req *personal_schedule.UpsertWorkRequest) error {
// 	if err := parseOptionalID(req.StatusId); err != nil {
// 		return fmt.Errorf("invalid StatusId: %v", err)
// 	}
// 	if err := parseOptionalID(req.DifficultyId); err != nil {
// 		return fmt.Errorf("invalid DifficultyId: %v", err)
// 	}
// 	if err := parseOptionalID(req.PriorityId); err != nil {
// 		return fmt.Errorf("invalid PriorityId: %v", err)
// 	}
// 	if err := parseOptionalID(req.TypeId); err != nil {
// 		return fmt.Errorf("invalid TypeId: %v", err)
// 	}
// 	if err := parseOptionalID(req.CategoryId); err != nil {
// 		return fmt.Errorf("invalid CategoryId: %v", err)
// 	}
// 	if err := parseOptionalID(req.GoalId); err != nil {
// 		return fmt.Errorf("invalid GoalId: %v", err)
// 	}
	
// 	for _, notificationID := range req.NotificationIds {
// 		_, err := bson.ObjectIDFromHex(notificationID)
// 		if err != nil {
// 			return fmt.Errorf("invalid NotificationId %s: %v", notificationID, err)
// 		}
// 	}

// 	for _, task := range req.SubTasks {
// 		if task.Id != nil && *task.Id != "" {
// 			if err := parseOptionalID(*task.Id); err != nil {
// 				return fmt.Errorf("invalid SubTask Id %s: %v", *task.Id, err)
// 			}
// 		}
// 	}

// 	if req.Id != nil && *req.Id != "" {
// 		if err := parseOptionalID(*req.Id); err != nil {
// 			return fmt.Errorf("invalid Work Id %s: %v", *req.Id, err)
// 		}

// 		workID, _ := bson.ObjectIDFromHex(*req.Id)

// 		existingWork, err := wv.workRepo.GetWorkByID(ctx, workID)
// 		if err != nil {
// 			return fmt.Errorf("error checking existing work: %v", err)
// 		}
// 		if existingWork == nil {
// 			return fmt.Errorf("work with Id %s does not exist", *req.Id)
// 		}
// 		if existingWork.UserID != req.UserId {
// 			return fmt.Errorf("work with Id %s does not belong to user %s", *req.Id, req.UserId)
// 		}
// 	}

// 	return nil
// }
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
	if _, err := bson.ObjectIDFromHex(req.GoalId); err != nil {
		return fmt.Errorf("invalid GoalId")
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