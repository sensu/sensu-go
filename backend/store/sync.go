package store

import (
	"context"
)

// Mutex is a static identifier for a particular mutex.
//
// In lieu of a central repository of mutex identities, use the four most
// significant bytes as a codebase identifier.
type Mutex int64

const (
	// BitmaskMutexOSS bitmask for sensu-go mutexes.
	BitmaskMutexOSS Mutex = 0
	// BitmaskMutexEE bitmask for mutexes in the enterprise build.
	BitmaskMutexEE Mutex = 0xee << 32
)

// Mutexes
const (
	// mutex for tessend telemetry
	MutexTelemetry Mutex = iota ^ BitmaskMutexOSS
)

// MutexHandler should listen for context cancellation. If a mutex is lost,
// a SynchronizedExecutor will signal the handler to wrap up this way.
type MutexHandler func(context.Context) error

type SynchronizedExecutor interface {
	// Execute MutexHandler once the mutex can be acquired.
	Execute(context.Context, Mutex, MutexHandler) error
}
