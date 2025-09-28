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
