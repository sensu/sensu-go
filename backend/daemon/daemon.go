package daemon

// A Daemon is a managed subprocess comprised of one or more goroutines that
// can be managed via a consistent, simple interface.
type Daemon interface {
	// Start starts the daemon, returning an error if preconditions for startup
	// fail.
	Start() error

	// Stop stops the daemon, returning an error if one was encountered during
	// shutdown.
	Stop() error

	// Status returns nil if the Daemon is healthy, otherwise it returns an error.
	Status() error

	// Err returns a channel that the caller can use to listen for terminal errors
	// indicating a premature shutdown of the Daemon.
	Err() <-chan error
}
