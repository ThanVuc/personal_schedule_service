package utils

import (
	"context"
	"personal_schedule_service/global"
	"personal_schedule_service/proto/common"

	"go.uber.org/zap"
)

var ErrorMessage = map[string]string{
	"DatabaseError": "Database operation failed",
	"NotFoundError": "Resource not found",
	"RuntimeError":  "An unexpected error occurred",
}

func DatabaseError(ctx context.Context, err error) *common.Error {
	logger := global.Logger
	requestId := GetRequestIDFromOutgoingContext(ctx)
	logger.Error("Database operation failed", requestId, zap.Error(err))
	e := &common.Error{
		Code:    common.ErrorCode_ERROR_CODE_DATABASE_ERROR,
		Message: "Database operation failed",
	}
	return e
}

func NotFoundError(ctx context.Context, err error) *common.Error {
	logger := global.Logger
	requestId := GetRequestIDFromOutgoingContext(ctx)
	logger.Error("Resource not found", requestId, zap.Error(err))
	e := &common.Error{
		Code:    common.ErrorCode_ERROR_CODE_NOT_FOUND,
		Message: "Not found error",
	}
	return e
}

func RuntimeError(ctx context.Context, err error) *common.Error {
	logger := global.Logger
	requestId := GetRequestIDFromOutgoingContext(ctx)
	logger.Error("An unexpected error occurred: runtime error", requestId, zap.Error(err))
	e := &common.Error{
		Code:    common.ErrorCode_ERROR_CODE_RUN_TIME_ERROR,
		Message: "An unexpected error occurred: runtime error",
	}
	return e
}

func UnauthorizedError(ctx context.Context, err error) *common.Error {
	logger := global.Logger
	requestId := GetRequestIDFromOutgoingContext(ctx)
	logger.Error("Unauthorized access", requestId, zap.Error(err))
	e := &common.Error{
		Code:    common.ErrorCode_ERROR_CODE_UNAUTHORIZED,
		Message: "Unauthorized access",
	}
	return e
}

func PermissionDeniedError(ctx context.Context, err error) *common.Error {
	logger := global.Logger
	requestId := GetRequestIDFromOutgoingContext(ctx)
	logger.Error("Permission denied", requestId, zap.Error(err))
	e := &common.Error{
		Code:    common.ErrorCode_ERROR_CODE_PERMISSION_DENIED,
		Message: "Permission denied",
	}
	return e
}

func InternalServerError(ctx context.Context, err error) *common.Error {
	logger := global.Logger
	requestId := GetRequestIDFromOutgoingContext(ctx)
	logger.Error("Internal server error", requestId, zap.Error(err))
	e := &common.Error{
		Code:    common.ErrorCode_ERROR_CODE_INTERNAL_ERROR,
		Message: "Internal server error",
	}
	return e
}

// CustomError allows creating a custom error with a specific error code and message.
// with code is the type of common.ErrorCode
// errorCode is an integer representing a specific error scenario within that type.
// errorCode is defined in the ErrorMessage const map on each service level.
func CustomError(ctx context.Context, code common.ErrorCode, errorCode int32, err error) *common.Error {
	logger := global.Logger
	requestId := GetRequestIDFromOutgoingContext(ctx)
	logger.Error("custom error occurred", requestId, zap.Error(err))
	e := &common.Error{
		Code:      code,
		Message:   err.Error(),
		ErrorCode: &errorCode,
	}
	return e
}

func GetErrorKeyByCode(code common.ErrorCode) string {
	// ERROR_CODE_UNAUTHORIZED = 0;
	// ERROR_CODE_NOT_FOUND = 1;
	// ERROR_CODE_DATABASE_ERROR = 2;
	// ERROR_CODE_RUN_TIME_ERROR = 3;
	// ERROR_CODE_PERMISSION_DENIED = 4;
	// ERROR_CODE_INTERNAL_ERROR = 5;
	switch code {
	case common.ErrorCode_ERROR_CODE_UNAUTHORIZED:
		return "UnauthorizedError"
	case common.ErrorCode_ERROR_CODE_NOT_FOUND:
		return "NotFoundError"
	case common.ErrorCode_ERROR_CODE_DATABASE_ERROR:
		return "DatabaseError"
	case common.ErrorCode_ERROR_CODE_RUN_TIME_ERROR:
		return "RuntimeError"
	case common.ErrorCode_ERROR_CODE_PERMISSION_DENIED:
		return "PermissionDeniedError"
	case common.ErrorCode_ERROR_CODE_INTERNAL_ERROR:
		return "InternalServerError"
	default:
		return "RuntimeError"
	}
}
