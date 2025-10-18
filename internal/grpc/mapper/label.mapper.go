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
		labelProto := &personal_schedule.Label{
			Id:        label.ID.Hex(),
			Name:      label.Name,
			Color:     *label.Color,
			Key:       label.Key,
			Meaning:   *label.Meaning,
			Note:      *label.Note,
			LabelType: int32(label.LabelType),
		}
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
