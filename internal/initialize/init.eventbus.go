package initialize

import (
	"fmt"
	"personal_schedule_service/global"

	"github.com/thanvuc/go-core-lib/eventbus"
)

func initEventBus() error {
	rabbitMqConfig := global.Config.RabbitMQ
	uri := fmt.Sprintf(
		"amqp://%s:%s@%s:%d",
		rabbitMqConfig.User,
		rabbitMqConfig.Password,
		rabbitMqConfig.Host,
		rabbitMqConfig.Port,
	)

	connector, err := eventbus.NewConnector(
		uri,
		global.Logger,
	)

	if err != nil {
		panic(err)
	}

	global.EventBusConnector = connector

	return nil
}
