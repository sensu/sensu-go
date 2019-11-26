package agent

import (
	"context"
	"errors"
)

// StatsdUnsupported is returned when statsd can't be supported on the platform.
var StatsdUnsupported = errors.New("statsd not supported on this platform")

// StatsdServer is the interface the agent requires to run a statsd server.
type StatsdServer interface {
	Run(context.Context) error
}
