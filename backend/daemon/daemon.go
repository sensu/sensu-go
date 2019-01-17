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
