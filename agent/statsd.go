package agent

import (
	"context"
	"errors"
)

// ErrStatsdUnsupported is returned when statsd can't be supported on the platform.
var ErrStatsdUnsupported = errors.New("statsd not supported on this platform")

// DEPRECATED: use ErrStatsdUnsupported
var StatsdUnsupported = ErrStatsdUnsupported

// StatsdServer is the interface the agent requires to run a statsd server.
type StatsdServer interface {
	Run(context.Context) error
}
