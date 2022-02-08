package agent

import (
	"context"
)

// GetMetricsAddr always returns ""
func GetMetricsAddr(s StatsdServer) string {
	return ""
}

// statsdServer is a no-op statsd server for solaris support.
// the gostatsd package requires a library that can't be built on solaris.
type statsdServer struct{}

func (s statsdServer) Run(context.Context) error {
	return ErrStatsdUnsupported
}

func NewStatsdServer(*Agent) statsdServer {
	return statsdServer{}
}
