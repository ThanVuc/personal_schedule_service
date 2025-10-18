package services

import (
	"context"
	"personal_schedule_service/internal/grpc/helper"
	"personal_schedule_service/internal/grpc/mapper"
	"personal_schedule_service/internal/grpc/utils"
	"personal_schedule_service/internal/repos"
	app_error "personal_schedule_service/pkg/settings/error"
	"personal_schedule_service/proto/common"
	"personal_schedule_service/proto/personal_schedule"

	"github.com/thanvuc/go-core-lib/log"
	"go.uber.org/zap"
)

type labelService struct {
	labelHelper helper.LabelHelper
	logger      log.Logger
	labelRepo   repos.LabelRepo
	labelMapper mapper.LabelMapper
}

func (s *labelService) SeedLabels(ctx context.Context) error {
	count, err := s.labelRepo.CountLabels(ctx)
	if err != nil {
		s.logger.Error("Cannot get labels", "", zap.Error(err))
		return err
	}

	if count > 0 {
		s.logger.Info("Labels already initialized, skip inserting labels", "")
		return nil
	}

	labels := s.labelHelper.GenerateLabel()
	err = s.labelRepo.InsertLabels(ctx, &labels)
	if err != nil {
		s.logger.Error("Cannot Initalize Labels", "", zap.Error(err))
		panic("Cannot Initalize Labels")
	}
	s.logger.Info("Labels initialized successfully", "")
	return nil
}

func (s *labelService) GetLabelPerTypes(ctx context.Context, req *common.EmptyRequest) (*personal_schedule.GetLabelPerTypesResponse, error) {
	labels, err := s.labelRepo.GetLabels(ctx)
	if err != nil {
		return &personal_schedule.GetLabelPerTypesResponse{
			LabelPerTypes: nil,
			// Specific error func without return specific error code to frontend
			Error: utils.DatabaseError(ctx, err),
		}, err
	}

	if len(labels) == 0 {
		return &personal_schedule.GetLabelPerTypesResponse{
			LabelPerTypes: nil,
			// Custom error func with specific return code to frontend
			Error: utils.CustomError(ctx, common.ErrorCode_ERROR_CODE_NOT_FOUND, app_error.LabelNotFoundCode, err),
		}, nil
	}

	labelPerTypes := s.labelMapper.MapLabelsToLabelTypesProto(labels)

	resp := &personal_schedule.GetLabelPerTypesResponse{
		LabelPerTypes: labelPerTypes,
		Error:         nil,
	}

	return resp, nil
}
