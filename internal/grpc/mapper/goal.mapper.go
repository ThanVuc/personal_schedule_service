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
	}
}

func (m *goalMapper) mapLabelsToProto(labels []collection.Label) []*personal_schedule.Label {
	protoLabels := make([]*personal_schedule.Label, 0, len(labels))
	for _, label := range labels {
		lc := ""
		if label.Color != nil {
			lc = *label.Color
		}
		protoLabels = append(protoLabels, &personal_schedule.Label{
			Id:    label.ID.Hex(),
			Name:  label.Name,
			Color: lc,
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
