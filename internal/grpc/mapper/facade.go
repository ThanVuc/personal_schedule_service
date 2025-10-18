package mapper

import (
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/proto/personal_schedule"
)

type (
	LabelMapper interface {
		MapLabelsToLabelTypesProto(labels []collection.Label) []*personal_schedule.LabelPerType
	}
)

func NewLabelMapper() LabelMapper {
	return &labelMapper{}
}
