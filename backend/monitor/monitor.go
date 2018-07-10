package monitor

import (
	"time"

	"github.com/sensu/sensu-go/types"
)

// Interface is the monitor interface.
type Interface interface {
	// Stop stops the monitor.
	Stop()

	// IsStopped returns true if the monitor is stopped, false otherwise.
	IsStopped() bool

	// HandleUpdate handles an update event with the monitor's UpdateHandler.
	HandleUpdate(event *types.Event) error

	// HandleFailure handles a failure event with the monitor's FailureHandler.
	HandleFailure(entity *types.Entity, event *types.Event) error
	GetTimeout() time.Duration
}

// Handler is a handler for monitor events.
type Handler interface {
	FailureHandler
	ErrorHandler
}

// FailureHandler handles monitoring failures.
type FailureHandler interface {
	HandleFailure(event *types.Event) error
}

// ErrorHandler handles internal monitor errors.
type ErrorHandler interface {
	HandleError(error)
}

// ErrorHandlerFunc implements ErrorHandler
type ErrorHandlerFunc func(error)

func (e ErrorHandlerFunc) HandleError(err error) {
	e(err)
}
