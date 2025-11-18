package mapper

import (
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/internal/repos"
	"personal_schedule_service/proto/personal_schedule"
)

type (
	LabelMapper interface {
		MapLabelsToLabelTypesProto(labels []collection.Label) []*personal_schedule.LabelPerType
		MapLabelsToLabelsProto(labels []collection.Label) []*personal_schedule.Label
	}

	GoalMapper interface {
		ConvertAggregatedGoalsToProto(aggGoals []repos.AggregatedGoal) []*personal_schedule.Goal
		MapUpsertProtoToModels(req *personal_schedule.UpsertGoalRequest) (*collection.Goal, []collection.GoalTask, error)
		MapAggregatedToDetailProto(aggGoal repos.AggregatedGoal, dbTasks []collection.GoalTask) *personal_schedule.GoalDetail
	}

	WorkMapper interface {
		MapUpsertProtoToModels(req *personal_schedule.UpsertWorkRequest) (*collection.Work, []collection.SubTask, error)
	}
)

func NewLabelMapper() LabelMapper {
	return &labelMapper{}
}

func NewGoalMapper() GoalMapper {
	return &goalMapper{}
}

func NewWorkMapper() WorkMapper {
	return &workMapper{}
}