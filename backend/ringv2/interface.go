package ringv2

import (
	"context"

	"github.com/robfig/cron"
)

// Interface specifies the interface of a ring.
type Interface interface {
	// Add adds a value to the ring.
	Add(ctx context.Context, value string, keepalive int64) error

	// Remove removes a value from the ring.
	Remove(context.Context, string) error

	// Watch returns a channel that will receive all events that occur in the
	// ring. The events will be one of the defined EventTypes. When the given
	// context is canceled, the watcher will terminate.
	//
	// The items parameter controls the number of values that will be in the
	// events. If there are fewer items in the ring than the requested items,
	// then the events will contain repetitions in order to satisfy the number
	// of items requested.
	Watch(ctx context.Context, items int) <-chan Event

	// SetInterval sets the ring's interval. After setting the interval,
	// future trigger events will occur at least the interval seconds apart.
	SetInterval(ctx context.Context, seconds int64) error

	// SetCron sets the cron schedule. It causes SetInterval to have no effect.
	SetCron(cron.Schedule)
}
