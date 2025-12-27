//go:build wireinject

package wire

import (
	"personal_schedule_service/internal/eventbus/handler"
	"personal_schedule_service/internal/repos"

	"github.com/google/wire"
)

func InjectSyncAuthHandler() *handler.SyncAuthHandler {
	wire.Build(
		repos.NewUserRepo,
		handler.NewSyncAuthHandler,
	)

	return nil
}
