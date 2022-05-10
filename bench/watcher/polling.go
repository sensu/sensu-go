package watcher

import (
	"context"
	"time"
)

type Action uint

const (
	Create Action = iota
	Update
	Delete
	WatchError
)

type Event struct {
	Action   Action
	Resource interface{}
}

type Wrapper struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Id        string
	Resource  interface{}
}

type Watcher struct {
	// Interval to call Updates for updates
	Interval time.Duration
	// TxnWindow the amount of overlap when querying for updates
	// must be less than Interval
	TxnWindow time.Duration
	// Head gets the current most recent updated at from the collection
	// e.g. SELECT updated_at from mytable ORDER BY updated_at DESC LIMIT 1;
	Head func(context.Context) (time.Time, error)
	// Updates provies a list of updates that have occured since a time (inclusive)
	Updates func(context.Context, time.Time) ([]Wrapper, error)
}

func (w *Watcher) Watch(ctx context.Context) (<-chan Event, error) {
	e := make(chan Event, 128)
	start, err := w.Head(ctx)
	if err != nil {
		return e, err
	}
	go w.watch(ctx, start, e)
	return e, nil
}

func (w *Watcher) watch(ctx context.Context, start time.Time, events chan<- Event) {
	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()
	// collection of notifications delivered
	notifications := make(map[string]time.Time)

	var prevMostRecent, mostRecent time.Time

	for {
		select {
		case <-ticker.C:
			updates, err := w.Updates(ctx, start)
			if err != nil {
				continue
			}
			for _, update := range updates {
				if prevTs, ok := notifications[update.Id]; ok && prevTs == update.UpdatedAt {
					continue
				}
				notifications[update.Id] = update.UpdatedAt
				if update.UpdatedAt.After(mostRecent) {
					mostRecent = update.UpdatedAt
				}

				action := Update
				if update.DeletedAt != nil {
					action = Delete
				} else if update.CreatedAt == update.UpdatedAt {
					action = Create
				}
				events <- Event{Action: action, Resource: update.Resource}
			}
			// prepare for next interval
			if mostRecent.After(prevMostRecent) {
				start = mostRecent.Add(-1 * w.TxnWindow)
			} else {
				start = mostRecent
			}
			prevMostRecent = mostRecent
			for id, ts := range notifications {
				if ts.Before(start) {
					delete(notifications, id)
				}
			}
		case <-ctx.Done():
			close(events)
			return
		}
	}
}
