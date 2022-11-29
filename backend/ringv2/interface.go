package ringv2

import (
	"context"
	"errors"
	"strings"

	"github.com/robfig/cron/v3"
	"github.com/sensu/sensu-go/backend/store"
)

type deleteEntityContextKeyT struct{}

// DeleteEntityContextKey can be set to tell the ring implementation to delete
// the entity as well as the entity's ring association. Does not currently apply
// to the etcd-based ring.
var DeleteEntityContextKey = deleteEntityContextKeyT{}

// DeleteEntityContext modifies a context with a value that can inform ring
// implementations that deleting the entity is desired as well as deleting
// the ring association. Does not currently apply to the etcd-based ring.
func DeleteEntityContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, DeleteEntityContextKey, struct{}{})
}

// Event represents an event that occurred in a ring. The event can originate
// from any ring client.
type Event struct {
	// Type is the type of the event.
	Type EventType

	// Values are the ring items associated with the event. For trigger events,
	// the length of Values will be equal to the results per interval.
	Values []string

	// Err is any error that occurred while processing the event.
	Err error
}

// EventType is an enum that describes the type of event received by watchers.
type EventType int

const (
	// EventError is a message sent when a ring processing error occurs.
	EventError EventType = iota
	// EventAdd is a message sent when an item is added to the ring.
	EventAdd
	// EventRemove is a message sent when an item is removed from the ring.
	EventRemove
	// EventTrigger is a message sent when a ring item has moved from the front of the queue to the back.
	EventTrigger
	// EventClosing is a message sent when the ring is closing due to context cancellation.
	EventClosing
)

func (e EventType) String() string {
	switch e {
	case EventAdd:
		return "EventAdd"
	case EventRemove:
		return "EventRemove"
	case EventTrigger:
		return "EventTrigger"
	case EventError:
		return "EventError"
	case EventClosing:
		return "EventClosing"
	default:
		return "INVALID"
	}
}

// Interface is the interface of a round-robin ring.
type Interface interface {
	// Subscribe receives one or more items from the ring, on a schedule described
	// by the Subscription. If the Subscription is not valid, Subscribe panics.
	Subscribe(ctx context.Context, sub Subscription) <-chan Event

	// Remove removes an item from the ring.
	Remove(ctx context.Context, value string) error

	// Add adds an item to the ring, with a keepalive in seconds. After the
	// first Add, Add must be called before the keepalive expires, or the ring
	// item will be removed from the ring.
	Add(ctx context.Context, value string, keepalive int64) error

	// IsEmpty returns true if the ring is empty.
	IsEmpty(ctx context.Context) (bool, error)
}

// Subscription is configuration for the Subscribe method of a ring.
type Subscription struct {
	// Name is the unique name of this subscription. Typically, it would be the
	// name of the check, in a check scheduling scenario.
	Name string

	// Items is the number of ring items to get from the ring in a given
	// scheduling iteration. This is useful when doing proxy check round robin
	// scheduling, with splay, as it requires several ring items at a time.
	Items int

	// IntervalSchedule is the number of seconds to wait between receiving ring
	// items. If set, CronSchedule must not be set.
	IntervalSchedule int

	// CronSchedule is the cron schedule to use for waiting to receive ring
	// items. If set, IntervalSchedule must not be set.
	CronSchedule string
}

// Validate returns an error if the subscription is improperly configured.
//
// A well-configured Subscription has a non-empty Name, a
// non-zero Items, and one of IntervalSchedule or CronSchedule defined.
func (r Subscription) Validate() error {
	if r.Name == "" {
		return errors.New("ring: check subscription is empty")
	}
	if r.Items <= 0 {
		return errors.New("ring: number of items in subscription <= 0")
	}
	if r.IntervalSchedule == 0 && r.CronSchedule == "" {
		return errors.New("ring: no schedule defined")
	}
	if r.IntervalSchedule != 0 && r.CronSchedule != "" {
		return errors.New("ring: both cron and interval schedules defined")
	}
	if r.CronSchedule != "" {
		if _, err := cron.ParseStandard(r.CronSchedule); err != nil {
			return err
		}
	}
	return nil
}

// Path returns the canonical path to a ring.
func Path(namespace, subscription string) string {
	return store.NewKeyBuilder("rings").WithNamespace(namespace).Build(subscription)
}

// UnPath parses a path created by Path.
func UnPath(key string) (namespace, subscription string, err error) {
	parts := strings.Split(key, "/")
	if len(parts) < 5 {
		return "", "", errors.New("invalid ring key")
	}
	return parts[3], parts[4], nil
}
