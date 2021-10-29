package daemon

import "context"

// A Daemon is a managed subprocess comprised of one or more goroutines that
// can be managed via a consistent, simple interface.
type Daemon interface {
	// Start starts the daemon, returning an error if preconditions for startup
	// fail. The daemon will run until the provided context is cancelled. The
	// context error will be returned if no other error is encountered first.
	Start(context.Context) error

	// Stop waits until the daemon has come to a stop. The context provided to
	// Start must be cancelled for this function to return.
	Stop() error

	// Err returns a channel that the caller can use to listen for terminal errors
	// indicating a premature shutdown of the Daemon.
	Err() <-chan error

	// Name returns the name of the daemon
	Name() string
}

// Get returns the daemon with the provided name
func Get(daemons []Daemon, name string) Daemon {
	for _, daemon := range daemons {
		if daemon.Name() == name {
			return daemon
		}
	}
	return nil
}
