package mapper

import (
	"personal_schedule_service/internal/collection"
	"personal_schedule_service/proto/personal_schedule"
	"sort"
)

type labelMapper struct{}

func (m *labelMapper) MapLabelsToLabelTypesProto(labels []collection.Label) []*personal_schedule.LabelPerType {
	var protoLabels []*personal_schedule.LabelPerType
	var labelTypeMap = make(map[int32][]*personal_schedule.Label)
	for _, label := range labels {
		labelProto := m.MapLabelToLabelProto(label)
		labelTypeMap[int32(label.LabelType)] = append(labelTypeMap[int32(label.LabelType)], labelProto)
	}

	for labelType, labelList := range labelTypeMap {
		labelTypeProto := &personal_schedule.LabelPerType{
			Type:   labelType,
			Labels: labelList,
		}
		protoLabels = append(protoLabels, labelTypeProto)
	}

	sort.Slice(protoLabels, func(i, j int) bool {
		return protoLabels[i].Type < protoLabels[j].Type
	})

	return protoLabels
}

func (m *labelMapper) MapLabelsToLabelsProto(labels []collection.Label) []*personal_schedule.Label {
	var protoLabels []*personal_schedule.Label
	for _, label := range labels {
		labelProto := m.MapLabelToLabelProto(label)
		protoLabels = append(protoLabels, labelProto)
	}

	return protoLabels
}

func (m *labelMapper) MapLabelToLabelProto(labels collection.Label) *personal_schedule.Label {
	return &personal_schedule.Label{
		Id:        labels.ID.Hex(),
		Name:      labels.Name,
		Color:     *labels.Color,
		Key:       labels.Key,
		Meaning:   *labels.Meaning,
		Note:      *labels.Note,
		LabelType: int32(labels.LabelType),
	}
}
