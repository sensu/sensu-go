package gostatsd

import (
	"context"

	"github.com/spf13/viper"
)

// BackendFactory is a function that returns a Backend.
type BackendFactory func(*viper.Viper) (Backend, error)

// SendCallback is called by Backend.SendMetricsAsync() to notify about the result of operation.
// A list of errors is passed to the callback. It may be empty or contain nil values. Every non-nil value is an error
// that happened while sending metrics.
type SendCallback func([]error)

// Backend represents a backend.
type Backend interface {
	// Name returns the name of the backend.
	Name() string
	// SendMetricsAsync flushes the metrics to the backend, preparing payload synchronously but doing the send asynchronously.
	// Must not read/write MetricMap asynchronously.
	SendMetricsAsync(context.Context, *MetricMap, SendCallback)
	// SendEvent sends event to the backend.
	SendEvent(context.Context, *Event) error
}

// RunnableBackend represents a backend that needs a Run method to be executed to work.
type RunnableBackend interface {
	Backend
	// Run executes backend send operations. Should be started in a goroutine.
	Run(context.Context)
}
