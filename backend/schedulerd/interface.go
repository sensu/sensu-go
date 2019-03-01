package schedulerd

import corev2 "github.com/sensu/sensu-go/api/core/v2"

// Scheduler is the abstract interface of a scheduler.
type Scheduler interface {
	// Start stops the scheduler.
	Start()

	// Stop stops the scheduler.
	Stop() error

	// Interrupt refreshes the state of the scheduler.
	Interrupt(*corev2.CheckConfig)

	// Type returns the scheduler type
	Type() SchedulerType
}

type SchedulerType int

const (
	IntervalType SchedulerType = iota
	CronType
	RoundRobinIntervalType
	RoundRobinCronType
)

func (s SchedulerType) String() string {
	switch s {
	case IntervalType:
		return "interval"
	case CronType:
		return "cron"
	case RoundRobinIntervalType:
		return "round-robin interval"
	case RoundRobinCronType:
		return "round-robin cron"
	default:
		return "invalid"
	}
}

// GetSchedulerType gets the SchedulerType for a given check config.
func GetSchedulerType(check *corev2.CheckConfig) SchedulerType {
	if check.Cron != "" {
		if check.RoundRobin {
			return RoundRobinCronType
		}
		return CronType
	}
	if check.RoundRobin {
		return RoundRobinIntervalType
	}
	return IntervalType
}
