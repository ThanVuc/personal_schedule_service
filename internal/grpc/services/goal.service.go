package services

import (
	"context"
	"personal_schedule_service/internal/grpc/mapper"
	"personal_schedule_service/internal/grpc/utils"
	"personal_schedule_service/internal/repos"
	"personal_schedule_service/proto/personal_schedule"

	"github.com/thanvuc/go-core-lib/log"
	"go.uber.org/zap"
)

type goalService struct {
	logger     log.Logger
	goalRepo   repos.GoalRepo
	goalMapper mapper.GoalMapper
}

func (s *goalService) GetGoals(ctx context.Context, req *personal_schedule.GetGoalsRequest) (*personal_schedule.GetGoalsResponse, error) {
	goals, totalGoals, err := s.goalRepo.GetGoals(ctx, req)

	if err != nil {
		s.logger.Error("Error fetching goals from repo", "err", zap.Error(err))
		return &personal_schedule.GetGoalsResponse{
			Error:    utils.DatabaseError(ctx, err),
			Goals:    nil,
			PageInfo: utils.ToPageInfo(req.PageQuery.Page, req.PageQuery.PageSize, int32(totalGoals)),
		}, err
	}

	if totalGoals == 0 {
		return &personal_schedule.GetGoalsResponse{
			Error:      nil,
			Goals:      nil,
			TotalGoals: 0,
			PageInfo:   utils.ToPageInfo(req.PageQuery.Page, req.PageQuery.PageSize, int32(totalGoals)),
		}, nil
	}

	protoGoals := s.goalMapper.ConvertAggregatedGoalsToProto(goals)

	pageInfo := utils.ToPageInfo(req.PageQuery.Page, req.PageQuery.PageSize, int32(totalGoals))

	resp := &personal_schedule.GetGoalsResponse{
		Goals:      protoGoals,
		PageInfo:   pageInfo,
		TotalGoals: int32(totalGoals),
		Error:      nil,
	}

	return resp, nil
}
