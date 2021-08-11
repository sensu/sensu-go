package ringv2

import (
	"context"
	"errors"

	"github.com/robfig/cron/v3"
)

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
