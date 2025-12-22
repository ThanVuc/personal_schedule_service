package eventbus_constant

import "github.com/thanvuc/go-core-lib/eventbus"

// base
const (
	VIET_NAM_JOB = "viet_nam_job"
	SERVICE      = "personal_schedule"
)

// exchange, queue, routing key
const (
	EXCHANGE = ".exchange"
)

// exchanges full names
const (
	VIET_NAM_JOB_EXCHANGE eventbus.ExchangeName = VIET_NAM_JOB + EXCHANGE
)

// routing keys full names
const (
	ONE_DAY_JOB_ROUTING_KEY string = VIET_NAM_JOB + ".one_day"
)

// queues full names
const (
	GIVE_UP_JOB_QUEUE  string = VIET_NAM_JOB + SERVICE + ".give_up_job.queue"
	REPEATED_JOB_QUEUE string = VIET_NAM_JOB + SERVICE + ".repeated_job.queue"
)
