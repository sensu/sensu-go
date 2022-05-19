package poll

import (
	"context"
	"fmt"
	"reflect"
	"time"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// Table interface for supporting poll-based watchers.
type Table interface {
	// Now returns the first timestamp for which Since currently would return zero records.
	Now(context.Context) (time.Time, error)
	// Since returns records that have changed at (inclusive) or since the timestamp.
	Since(context.Context, time.Time) ([]Row, error)
}

type Row struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Id        string
	Resource  corev3.Resource
}

// Poller configuration
type Poller struct {
	// Interval the poller uses to query the Table for updates.
	Interval time.Duration
	// TxnWindow is the duration that the poller caches and requeries
	// updates for. A zero or too aggresive TxnWindow will result in
	// skipped updates. An overly long TxnWindow will result in higher
	// memory usage and potentially extended poll intervals.
	TxnWindow time.Duration
	// Table implements the access methods required by the poller.
	Table Table
}

// Watch starts the poller that will publish WatchEvents to the channel.
// The channel will be closed after a WatchError or context timeout.
func (p *Poller) Watch(ctx context.Context, events chan storev2.WatchEvent) {
	defer close(events)

	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()

	start, err := p.Table.Now(ctx)
	if err != nil {
		events <- storev2.WatchEvent{Action: storev2.WatchError}
		return
	}

	cache := make(map[string]Row)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			updates, err := p.Table.Since(ctx, start)
			if err != nil {
				fmt.Println(err)
				events <- storev2.WatchEvent{Action: storev2.WatchError}
				return
			}
			newUpdates := merge(updates, cache)
			for _, update := range newUpdates {
				action := storev2.Update
				switch {
				case update.DeletedAt != nil:
					action = storev2.Delete
				case update.CreatedAt == update.UpdatedAt:
					action = storev2.Create
				}
				events <- storev2.WatchEvent{Action: action, Resource: update.Resource}
			}
		}
		// find next start lower bound
		latest := start
		for _, row := range cache {
			if row.UpdatedAt.After(latest) {
				latest = row.UpdatedAt
			}
		}
		next := latest.Add(-1 * p.TxnWindow)
		if next.After(start) {
			start = next
			// retire outdated cache entries
			for id, row := range cache {
				if next.After(row.UpdatedAt) {
					delete(cache, id)
				}
			}
		}
	}
}

func merge(rows []Row, cache map[string]Row) []Row {
	if len(rows) == 0 {
		return rows
	}
	results := make([]Row, 0, len(rows))
	for _, row := range rows {
		prev, ok := cache[row.Id]
		cache[row.Id] = row
		if !ok {
			results = append(results, row)
			continue
		}
		if row.UpdatedAt != prev.UpdatedAt ||
			row.CreatedAt != prev.CreatedAt ||
			row.DeletedAt != prev.DeletedAt {
			results = append(results, row)
			continue
		}
		// corner case when multiple transactions update a record at the same time
		if !reflect.DeepEqual(row.Resource, prev.Resource) {
			results = append(results, row)
			continue
		}
	}
	return results
}
