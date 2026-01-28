package cmd

import (
	"fmt"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/initialize"

	"github.com/thanvuc/go-core-lib/eventbus"
)

func Init() {
	initialize.InitConfigAndResources()
}

func RunConsole() {
	Init()

	publisher := eventbus.NewPublisher(
		global.EventBusConnector,
		"schedule_mcp_generate_work_exchange",
		eventbus.ExchangeTypeDirect,
		nil,
		nil,
		false,
	)

	for {
		// type the message in console to publish the message
		var message string
		println("Enter message to publish (type 'exit' to quit):")
		_, err := fmt.Scanln(&message)
		if err != nil {
			println("Error reading input:", err.Error())
			continue
		}
		if message == "exit" {
			break
		}
		err = publisher.Publish(
			nil,
			"console_publisher",
			[]string{"schedule_mcp_generate_work"},
			[]byte(message),
			nil,
		)
		if err != nil {
			println("Failed to publish message:", err.Error())
		} else {
			println("Message published:", message)
		}
	}
}
