package schedulerd

import (
	corev2 "github.com/sensu/core/v2"
)

// NoopScheduler does not schedule checks
// but serves as a placeholder
type NoopScheduler SchedulerType

func NewNoopScheduler(typ SchedulerType) NoopScheduler {
	return NoopScheduler(typ)
}

// Start starts the scheduler.
func (s NoopScheduler) Start() {
}

// Interrupt nothing
func (s NoopScheduler) Interrupt(check *corev2.CheckConfig) {
}

// Stop Nothing
func (s NoopScheduler) Stop() error {
	return nil
}

// Type returns underlying type of the NoopScheduler
func (s NoopScheduler) Type() SchedulerType {
	return SchedulerType(s)
}
