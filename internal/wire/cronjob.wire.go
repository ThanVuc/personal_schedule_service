//go:build wireinject

package wire

import (
	"personal_schedule_service/internal/cronjob/cronjob"
	"personal_schedule_service/internal/grpc/mapper"
	"personal_schedule_service/internal/grpc/services"
	"personal_schedule_service/internal/grpc/validation"
	"personal_schedule_service/internal/repos"

	"github.com/google/wire"
)

func InjectWorkCronJob() *cronjob.WorkCronJob {
	wire.Build(
		repos.NewWorkRepo,
		repos.NewLabelRepo,
		mapper.NewWorkMapper,
		validation.NewWorkValidator,
		services.NewWorkService,
		cronjob.NewWorkCronJob,
	)

	return nil
}
