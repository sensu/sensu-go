package ringv2

import "context"

// Interface specifies the interface of a ring.
type Interface interface {
	// Add adds a value to the ring.
	Add(context.Context, string) error

	// Remove removes a value from the ring.
	Remove(context.Context, string) error

	// Watch returns a channel that will receive all events that occur in the
	// ring. The events will be one of the defined EventTypes. When the given
	// context is canceled, the watcher will terminate.
	Watch(context.Context) <-chan Event

	// SetInterval sets the ring's interval. After setting the interval,
	// future trigger events will occur at least the interval seconds apart.
	SetInterval(ctx context.Context, seconds int64) error
}
