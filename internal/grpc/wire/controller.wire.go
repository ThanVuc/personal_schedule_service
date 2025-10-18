//go:build wireinject

package wire

import (
	"personal_schedule_service/internal/grpc/controller"
	"personal_schedule_service/internal/grpc/mapper"
	"personal_schedule_service/internal/grpc/services"
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
