package notifications_constant

import "github.com/thanvuc/go-core-lib/eventbus"

// base
const (
	SCHEDULED_NOTIFICATION = "scheduled_notification"
)

// exchange, queue, routing key
const (
	EXCHANGE    = ".exchange"
	QUEUE       = ".queue"
	ROUTING_KEY = ".routing_key"
)

// exchanges full names
const (
	NOTIFICATION_EXCHANGE eventbus.ExchangeName = SCHEDULED_NOTIFICATION + EXCHANGE
)

// queues full names
const (
	NOTIFICATION_QUEUE string = SCHEDULED_NOTIFICATION + QUEUE
)

// routing keys full names
const (
	NOTIFICATION_ROUTING_KEY string = SCHEDULED_NOTIFICATION + ROUTING_KEY
)
