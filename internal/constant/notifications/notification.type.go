package notifications_constant

import "github.com/thanvuc/go-core-lib/eventbus"

// service
const (
	NOTIFICATION_SERVICE = "notification"
)

// base
const (
	SCHEDULED_NOTIFICATION = "_scheduled_notification"
)

// exchange, queue, routing key
const (
	EXCHANGE    = "_exchange"
	QUEUE       = "_queue"
	ROUTING_KEY = "_routing_key"
)

// exchanges full names
const (
	NOTIFICATION_EXCHANGE eventbus.ExchangeName = NOTIFICATION_SERVICE + SCHEDULED_NOTIFICATION + EXCHANGE
)

// queues full names
const (
	NOTIFICATION_QUEUE string = NOTIFICATION_SERVICE + SCHEDULED_NOTIFICATION + QUEUE
)

// routing keys full names
const (
	NOTIFICATION_ROUTING_KEY string = NOTIFICATION_SERVICE + SCHEDULED_NOTIFICATION + ROUTING_KEY
)
