//go:build wireinject

package wire

import (
	"personal_schedule_service/internal/grpc/controller"
	"personal_schedule_service/internal/grpc/mapper"
	"personal_schedule_service/internal/grpc/services"
	"personal_schedule_service/internal/grpc/validation"
	"personal_schedule_service/internal/repos"

	"github.com/google/wire"
)

func InjectLabelController() *controller.LabelController {
	wire.Build(
		repos.NewLabelRepo,
		mapper.NewLabelMapper,
		services.NewLabelService,
		controller.NewLabelController,
	)

	return nil
}

func InjectGoalController() *controller.GoalController {
	wire.Build(
		repos.NewGoalRepo,
		mapper.NewGoalMapper,
		services.NewGoalService,
		controller.NewGoalController,
	)

	return nil
}

func InjectWorkController() *controller.WorkController {
	wire.Build(
		repos.NewWorkRepo,
		repos.NewLabelRepo,
		mapper.NewWorkMapper,
		services.NewWorkService,
		controller.NewWorkController,
		validation.NewWorkValidator,
	)
	return nil
}
