package helper

import "personal_schedule_service/internal/collection"

type (
	LabelHelper interface {
		GenerateLabel() []collection.Label
	}
)

func NewLabelHelper() LabelHelper {
	return &labelHelper{}
}
