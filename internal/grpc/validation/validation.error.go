package validation

import "personal_schedule_service/proto/common"

type ValidationError struct {
	Category common.ErrorCode
	Code     int32
	Message  string
}

func (ve *ValidationError) Error() string {
	return ve.Message
}

func NewValidationError(category common.ErrorCode, code int32, message string) error {
	return &ValidationError{
		Category: category,
		Code:     code,
		Message:  message,
	}
}
