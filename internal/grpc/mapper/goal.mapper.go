package mapper

import (
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/internal/repos"
	"personal_schedule_service/proto/personal_schedule"

	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type goalMapper struct{}

func (m *goalMapper) ConvertAggregatedGoalsToProto(aggGoals []repos.AggregatedGoal) []*personal_schedule.Goal {
	protoGoals := make([]*personal_schedule.Goal, 0, len(aggGoals))

	for _, aggGoal := range aggGoals {
		goalProto := m.MapAggregatedGoalToProto(aggGoal)
		protoGoals = append(protoGoals, goalProto)
	}
	return protoGoals
}

func (m *goalMapper) MapAggregatedGoalToProto(aggGoal repos.AggregatedGoal) *personal_schedule.Goal {
	sd := ""
	if aggGoal.ShortDescriptions != nil {
		sd = *aggGoal.ShortDescriptions
	}
	dd := ""
	if aggGoal.DetailedDescription != nil {
		dd = *aggGoal.DetailedDescription
	}
	stDate := ""
	if aggGoal.StartDate != nil {
		stDate = aggGoal.StartDate.Format(time.RFC3339)
	}
	enDate := ""
	if aggGoal.EndDate != nil {
		enDate = aggGoal.EndDate.Format(time.RFC3339)
	}

	return &personal_schedule.Goal{
		Id:                  aggGoal.ID.Hex(),
		Name:                aggGoal.Name,
		ShortDescriptions:   sd,
		DetailedDescription: dd,
		StartDate:           stDate,
		EndDate:             enDate,
		UserId:              aggGoal.UserID,
		Status:              m.mapLabelsToProto(aggGoal.Status),
		Difficulty:          m.mapLabelsToProto(aggGoal.Difficulty),
		Priority:            m.mapLabelsToProto(aggGoal.Priority),
		Worktype:            m.mapLabelsToProto(aggGoal.WorkLabel),
	}
}

func (m *goalMapper) mapLabelsToProto(labels []collection.Label) []*personal_schedule.LabelInfo {
	protoLabels := make([]*personal_schedule.LabelInfo, 0, len(labels))
	for _, label := range labels {
		lc := ""
		if label.Color != nil {
			lc = *label.Color
		}
		protoLabels = append(protoLabels, &personal_schedule.LabelInfo{
			Id:    label.ID.Hex(),
			Name:  label.Name,
			Color: lc,
			Key:   label.Key,
			LabelType: int32(label.LabelType),
		})
	}
	return protoLabels
}

func (m *goalMapper) MapUpsertProtoToModels(req *personal_schedule.UpsertGoalRequest) (*collection.Goal, []collection.GoalTask, error) {
	goalDB, err := m.mapProtoGoalToDB(req)
	if err != nil {
		return nil, nil, err
	}
	tasksDB, err := m.mapGoalTasktoDB(req.Tasks)
	if err != nil {
		return nil, nil, err
	}
	return goalDB, tasksDB, nil
}

func (m *goalMapper) mapProtoGoalToDB(req *personal_schedule.UpsertGoalRequest) (*collection.Goal, error) {
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
	workLabelID, err := bson.ObjectIDFromHex(req.WorkTypeId)
	if err != nil {
		return nil, err
	}

	startDate := time.Unix(*req.StartDate, 0)

	enđate := time.Unix(*req.EndDate, 0)

	return &collection.Goal{
		Name:                req.Name,
		ShortDescriptions:   req.ShortDescriptions,
		DetailedDescription: req.DetailedDescription,
		StartDate:           &startDate,
		EndDate:             &enđate,
		UserID:              req.UserId,
		StatusID:            statusID,
		DifficultyID:        difficultyID,
		PriorityID:          priorityID,
		WorkTypeID:          workLabelID,
	}, nil
}

func (m *goalMapper) mapGoalTasktoDB(pt []*personal_schedule.GoalTaskPayload) ([]collection.GoalTask, error) {
	taskDB := make([]collection.GoalTask, len(pt))

	for i, task := range pt {
		var taskId bson.ObjectID
		if task.Id != nil {
			taskId, _ = bson.ObjectIDFromHex(*task.Id)
		}
		taskDB[i] = collection.GoalTask{
			ID:          taskId,
			Name:        task.Name,
			IsCompleted: task.IsCompleted,
		}
	}
	return taskDB, nil
}

func (m *goalMapper) MapTasksToProto(dbTasks []collection.GoalTask) []*personal_schedule.GoalTaskPayload {
	protoTasks := make([]*personal_schedule.GoalTaskPayload, len(dbTasks))
	for i, task := range dbTasks {
		taskIDHex := task.ID.Hex()
		protoTasks[i] = &personal_schedule.GoalTaskPayload{
			Id:          &taskIDHex,
			Name:        task.Name,
			IsCompleted: task.IsCompleted,
		}
	}
	return protoTasks
}

func (m *goalMapper) MapAggregatedToDetailProto(aggGoal repos.AggregatedGoal, dbTasks []collection.GoalTask) *personal_schedule.GoalDetail {
	goalBaseProto := m.MapAggregatedGoalToProto(aggGoal)
	tasksProto := m.MapTasksToProto(dbTasks)
	return &personal_schedule.GoalDetail{
		Id:                  goalBaseProto.Id,
		Name:                goalBaseProto.Name,
		ShortDescriptions:   goalBaseProto.ShortDescriptions,
		DetailedDescription: goalBaseProto.DetailedDescription,
		StartDate:           goalBaseProto.StartDate,
		EndDate:             goalBaseProto.EndDate,
		Status:              goalBaseProto.Status,
		Difficulty:          goalBaseProto.Difficulty,
		Priority:            goalBaseProto.Priority,
		Worktype:            goalBaseProto.Worktype,
		UserId:              goalBaseProto.UserId,
		Tasks:               tasksProto,
	}
}
