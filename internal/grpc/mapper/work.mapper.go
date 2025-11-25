package mapper

import (
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/internal/repos"
	"personal_schedule_service/proto/personal_schedule"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type workMapper struct{}

// 4
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

// 5
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
		StartDate:           req.StartDate,
		EndDate:             req.EndDate,
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

// 6
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

// 1
func (m *workMapper) ConvertAggregatedWorksToProto(aggWorks []repos.AggregatedWork) []*personal_schedule.Work {
	protoWorks := make([]*personal_schedule.Work, 0, len(aggWorks))
	for _, aggWork := range aggWorks {
		protoWorks = append(protoWorks, m.MapAggregatedWorkToProto(aggWork))
	}
	return protoWorks
}

// 2
func (m *workMapper) MapAggregatedWorkToProto(aggWork repos.AggregatedWork) *personal_schedule.Work {
	sd := ""
	if aggWork.ShortDescriptions != nil {
		sd = *aggWork.ShortDescriptions
	}
	dd := ""
	if aggWork.DetailedDescription != nil {
		dd = *aggWork.DetailedDescription
	}

	return &personal_schedule.Work{
		Id:                  aggWork.ID.Hex(),
		Name:                aggWork.Name,
		ShortDescriptions:   &sd,
		DetailedDescription: &dd,
		StartDate:           *aggWork.StartDate,
		EndDate:             aggWork.EndDate,
		Labels: &personal_schedule.WorkLabelGroup{
			Status:     m.mapLabelsToProto(aggWork.Status),
			Difficulty: m.mapLabelsToProto(aggWork.Difficulty),
			Priority:   m.mapLabelsToProto(aggWork.Priority),
			Type:       m.mapLabelsToProto(aggWork.Type),
		},
		Category: m.mapLabelsToProto(aggWork.Category),
	}
}

// 3
func (m *workMapper) mapLabelsToProto(labels []collection.Label) []*personal_schedule.LabelInfo {
	protoLabels := make([]*personal_schedule.LabelInfo, 0, len(labels))
	for _, label := range labels {
		lc := ""
		if label.Color != nil {
			lc = *label.Color
		}
		protoLabels = append(protoLabels, &personal_schedule.LabelInfo{
			Id:        label.ID.Hex(),
			Name:      label.Name,
			Color:     lc,
			Key:       label.Key,
			LabelType: int32(label.LabelType),
		})
	}
	return protoLabels
}

func (m *workMapper) MapSubTasksToProto(subTasks []collection.SubTask) []*personal_schedule.SubTaskPayload {
	protoSubTasks := make([]*personal_schedule.SubTaskPayload, 0, len(subTasks))
	for _, subTask := range subTasks {
		idStr := subTask.ID.Hex()
		protoSubTasks = append(protoSubTasks, &personal_schedule.SubTaskPayload{
			Id:          &idStr,
			Name:        subTask.Name,
			IsCompleted: subTask.IsCompleted,
		})
	}
	return protoSubTasks
}

func (m *workMapper) MapAggregatedToWorkDetailProto(aggWork repos.AggregatedWork, subTasks []collection.SubTask) *personal_schedule.WorkDetail {
	workBaseProto := m.MapAggregatedWorkToProto(aggWork)
	subTasksProto := m.MapSubTasksToProto(subTasks)
	return &personal_schedule.WorkDetail{
		Id:                  workBaseProto.Id,
		Name:                workBaseProto.Name,
		ShortDescriptions:   workBaseProto.ShortDescriptions,
		DetailedDescription: workBaseProto.DetailedDescription,
		StartDate:           workBaseProto.StartDate,
		EndDate:             workBaseProto.EndDate,
		Labels: &personal_schedule.WorkLabelGroupDetail{
			Status:     workBaseProto.Labels.Status,
			Difficulty: workBaseProto.Labels.Difficulty,
			Priority:   workBaseProto.Labels.Priority,
			Type:       workBaseProto.Labels.Type,
			Category:   workBaseProto.Category,
		},
		SubTasks: subTasksProto,
	}
}
