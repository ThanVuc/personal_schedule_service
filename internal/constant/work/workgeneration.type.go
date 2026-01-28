package workgeneration_constant

import (
	"fmt"

	"github.com/thanvuc/go-core-lib/eventbus"
)

// Service === this repo service => consumer of the service
const (
	// Service
	SERVICE                   = "schedule_mcp"
	PERSONAL_SCHEDULE_SERVICE = "personal_schedule"
	NOTIFICATION_SERVICE      = "notification"

	// Feature
	WORK_GENERATION = "generate_work"
	WORK_TRANSFER   = "transfer_work"

	// Common
	EXCHANGE = "exchange"
	QUEUE    = "queue"
)

// Exchange
var (
	WORK_GENERATION_EXCHANGE eventbus.ExchangeName = eventbus.ExchangeName(fmt.Sprintf(
		"%s_%s_%s",
		SERVICE,
		WORK_GENERATION,
		EXCHANGE,
	))

	WORK_TRANSFER_EXCHANGE eventbus.ExchangeName = eventbus.ExchangeName(fmt.Sprintf(
		"%s_%s_%s",
		PERSONAL_SCHEDULE_SERVICE,
		WORK_TRANSFER,
		EXCHANGE,
	))

	NOTIFICATION_GENERATE_WORK_EXCHANGE eventbus.ExchangeName = eventbus.ExchangeName(fmt.Sprintf(
		"%s_%s_%s",
		NOTIFICATION_SERVICE,
		WORK_GENERATION,
		EXCHANGE,
	))
)

// Queue
var (
	WORK_GENERATION_QUEUE = fmt.Sprintf(
		"%s_%s_%s",
		SERVICE,
		WORK_GENERATION,
		QUEUE,
	)

	WORK_TRANSFER_QUEUE = fmt.Sprintf(
		"%s_%s_%s",
		PERSONAL_SCHEDULE_SERVICE,
		WORK_TRANSFER,
		QUEUE,
	)

	NOTIFICATION_GENERATE_WORK_QUEUE = fmt.Sprintf(
		"%s_%s_%s",
		NOTIFICATION_SERVICE,
		WORK_GENERATION,
		QUEUE,
	)
)

// Routing Key
var (
	WORK_GENERATION_ROUTING_KEY = fmt.Sprintf(
		"%s_%s",
		SERVICE,
		WORK_GENERATION,
	)

	WORK_TRANSFER_ROUTING_KEY = fmt.Sprintf(
		"%s_%s",
		PERSONAL_SCHEDULE_SERVICE,
		WORK_TRANSFER,
	)

	NOTIFICATION_GENERATE_WORK_ROUTING_KEY = fmt.Sprintf(
		"%s_%s",
		NOTIFICATION_SERVICE,
		WORK_GENERATION,
	)
)
