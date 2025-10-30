package mapper

import (
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/internal/repos"
	"personal_schedule_service/proto/personal_schedule"

	"time"
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
