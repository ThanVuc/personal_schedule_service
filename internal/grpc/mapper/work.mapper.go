// (File: mapper/work.mapper.go)
package mapper

import (
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/proto/personal_schedule"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type workMapper struct{}

func (m *workMapper) MapUpsertProtoToModels(req *personal_schedule.UpsertWorkRequest) (*collection.Work, []collection.SubTask, error) {
	goalDB, err := m.mapProtoWorkToDB(req)
	if err != nil {
		return nil, nil, err
	}
	tasksDB, err := m.mapSubTaskToDB(req.SubTasks)
	if err != nil {
		return nil, nil, err
	}
	return goalDB, tasksDB, nil
}

func (m *workMapper) mapProtoWorkToDB(req *personal_schedule.UpsertWorkRequest) (*collection.Work, error) {
	statusID, err := bson.ObjectIDFromHex(req.StatusId)
	if err != nil {
		return nil, err
	}
	difficultyID, err := bson.ObjectIDFromHex(req.DifficultyId)
	if err != nil {
		return nil, err
	}
	priorityID, err := bson.ObjectIDFromHex(req.PriorityId)
	if err != nil {
		return nil, err
	}
	typeID, err := bson.ObjectIDFromHex(req.TypeId)
	if err != nil {
		return nil, err
	}
	categoryID, err := bson.ObjectIDFromHex(req.CategoryId)
	if err != nil {
		return nil, err
	}
	goalID, err := bson.ObjectIDFromHex(req.GoalId)
	if err != nil {
		return nil, err
	}

	startDate := time.Unix(*req.StartDate, 0)

	endDate := time.Unix(req.EndDate, 0)

	notificationIDs := make([]bson.ObjectID, len(req.NotificationIds))
	for i, idStr := range req.NotificationIds {
		id, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			return nil, err
		}
		notificationIDs[i] = id
	}

	return &collection.Work{
		Name:                req.Name,
		ShortDescriptions:   req.ShortDescriptions,
		DetailedDescription: req.DetailedDescription,
		StartDate:           &startDate,
		EndDate:             endDate,
		NotificationIds:     notificationIDs,
		StatusID:            statusID,
		DifficultyID:        difficultyID,
		PriorityID:          priorityID,
		TypeID:              typeID,
		CategoryID:          categoryID,
		UserID:              req.UserId,
		GoalID:              goalID,
	}, nil
}

func (m *workMapper) mapSubTaskToDB(payload []*personal_schedule.SubTaskPayload) ([]collection.SubTask, error) {
	taskDB := make([]collection.SubTask, len(payload))
	for i, task := range payload {
		var taskID bson.ObjectID
		if task.Id != nil {
			taskID, _ = bson.ObjectIDFromHex(*task.Id)
		}

		taskDB[i] = collection.SubTask{
			ID:          taskID,
			Name:        task.Name,
			IsCompleted: task.IsCompleted,
		}
	}
	return taskDB, nil
}
