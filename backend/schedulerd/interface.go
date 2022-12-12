package schedulerd

import corev2 "github.com/sensu/core/v2"

// Scheduler is a check scheduler. It is responsible for determining the
// scheduling interval of a check, given a particular configuration.
// After Start(), the scheduler is active and will continue to schedule a
// check according to its schedule. When Interrupt is called, the schedule
// will be recalculated.
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

// SchedulerType represents the type of a scheduler.
type SchedulerType int

const (
	// IntervalType ...
	IntervalType SchedulerType = iota
	// CronType ...
	CronType
	// RoundRobinIntervalType ...
	RoundRobinIntervalType
	// RoundRobinCronType ...
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
