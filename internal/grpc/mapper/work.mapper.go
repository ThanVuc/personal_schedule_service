package mapper

import (
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/internal/grpc/utils"
	"personal_schedule_service/internal/repos"
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

	normalizedName := utils.RemoveAccent(req.Name)

	var goalID *bson.ObjectID
	if req.GoalId != nil && *req.GoalId != "" {
		id, err := bson.ObjectIDFromHex(*req.GoalId)
		if err != nil {
			return nil, err
		}
		goalID = &id
	}

	endDate := time.UnixMilli(req.EndDate)

	var startDate *time.Time
	if req.StartDate != nil {
		t := time.UnixMilli(*req.StartDate)
		startDate = &t
	}

	return &collection.Work{
		Name:                req.Name,
		ShortDescriptions:   req.ShortDescriptions,
		DetailedDescription: req.DetailedDescription,
		NameNormalized:      normalizedName,
		StartDate:           startDate,
		EndDate:             endDate,
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

func (m *workMapper) ConvertAggregatedWorksToProto(aggWorks []repos.AggregatedWork) []*personal_schedule.Work {
	protoWorks := make([]*personal_schedule.Work, 0, len(aggWorks))
	for _, aggWork := range aggWorks {
		protoWorks = append(protoWorks, m.MapAggregatedWorkToProto(aggWork))
	}
	return protoWorks
}

func (m *workMapper) MapAggregatedWorkToProto(aggWork repos.AggregatedWork) *personal_schedule.Work {
	sd := ""
	if aggWork.ShortDescriptions != nil {
		sd = *aggWork.ShortDescriptions
	}
	dd := ""
	if aggWork.DetailedDescription != nil {
		dd = *aggWork.DetailedDescription
	}

	var startdate, enddate int64
	if aggWork.StartDate != nil {
		startdate = aggWork.StartDate.UnixMilli()
	}
	enddate = aggWork.EndDate.UnixMilli()

	var goalProto *personal_schedule.GoalOfWork
	if len(aggWork.GoalInfo) > 0 {
		goalProto = &personal_schedule.GoalOfWork{
			Id:   aggWork.GoalInfo[0].ID.Hex(),
			Name: aggWork.GoalInfo[0].Name,
		}
	}

	return &personal_schedule.Work{
		Id:                  aggWork.ID.Hex(),
		Name:                aggWork.Name,
		ShortDescriptions:   &sd,
		DetailedDescription: &dd,
		StartDate:           startdate,
		EndDate:             enddate,
		Goal:                goalProto,
		Labels: &personal_schedule.WorkLabelGroup{
			Status:     m.mapLabelsToProto(aggWork.Status),
			Difficulty: m.mapLabelsToProto(aggWork.Difficulty),
			Priority:   m.mapLabelsToProto(aggWork.Priority),
			Type:       m.mapLabelsToProto(aggWork.Type),
			Draft:      m.mapLabelsToProto(aggWork.Draft),
		},
		Category: m.mapLabelsToProto(aggWork.Category),
		Overdue:  m.mapLabelsToProto(aggWork.Overdue),
	}
}

func (m *workMapper) mapLabelsToProto(labels []collection.Label) *personal_schedule.LabelInfo {
	if len(labels) == 0 {
		return nil
	}
	label := labels[0]
	lc := ""
	if label.Color != nil {
		lc = *label.Color
	}
	return &personal_schedule.LabelInfo{
		Id:        label.ID.Hex(),
		Name:      label.Name,
		Color:     lc,
		Key:       label.Key,
		LabelType: int32(label.LabelType),
	}
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
		Goal:                workBaseProto.Goal,
		Labels: &personal_schedule.WorkLabelGroupDetail{
			Status:     workBaseProto.Labels.Status,
			Difficulty: workBaseProto.Labels.Difficulty,
			Priority:   workBaseProto.Labels.Priority,
			Type:       workBaseProto.Labels.Type,
			Category:   workBaseProto.Category,
		},
		Draft:                 workBaseProto.Labels.Draft,
		SubTasks:              subTasksProto,
		RepeatSeriesStartDate: nil,
		RepeatSeriesEndDate:   nil,
	}
}
