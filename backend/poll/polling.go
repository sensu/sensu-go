package poll

import (
	"context"
	"fmt"
	"reflect"
	"time"
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
	Resource  interface{}
}

type RowChangeType uint

const (
	Create RowChangeType = iota
	Update
	Delete
)

type RowChange struct {
	Resource interface{}
	Change   RowChangeType
}

// Poller configuration
type Poller struct {
	// Interval the poller uses to query the Table for updates.
	Interval time.Duration
	// TxnWindow is the duration that the poller caches and requeries
	// updates for. A zero or too aggressive TxnWindow will result in
	// skipped updates. Longer TxnWindows lead to increased query
	// load on the underlying store.
	TxnWindow time.Duration
	// Table implements the access methods required by the poller.
	Table Table

	start    time.Time
	nextPoll time.Time
	cache    map[string]Row
}

// Initialize the Poller with the starting state of
// the collection.
// Initialize must be called exactly once before Next is called.
func (p *Poller) Initialize(ctx context.Context) error {
	var err error
	p.cache = make(map[string]Row)
	p.start, err = p.Table.Now(ctx)
	p.nextPoll = time.Now().Add(p.Interval)
	return err
}

// Next blocks until the next polling interval, then returns any changed rows.
func (p *Poller) Next(ctx context.Context) ([]RowChange, error) {
	nextInterval := time.NewTimer(time.Until(p.nextPoll))
	defer nextInterval.Stop()

	p.nextPoll = p.nextPoll.Add(p.Interval)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-nextInterval.C:
	}

	updates, err := p.Table.Since(ctx, p.start)
	if err != nil {
		return nil, fmt.Errorf("poller error getting updates since %v: %v", p.start, err)
	}
	newUpdates := merge(updates, p.cache)
	results := make([]RowChange, len(newUpdates))
	for i, update := range newUpdates {
		var change RowChangeType
		switch {
		case update.DeletedAt != nil:
			change = Delete
		case update.CreatedAt == update.UpdatedAt:
			change = Create
		default:
			change = Update
		}
		results[i].Resource = update.Resource
		results[i].Change = change
	}

	p.advanceStartTime()
	return results, nil
}

// advanceStartTime progresses the poll window and clears outdated
// cache entries.
func (p *Poller) advanceStartTime() {
	latest := p.start
	for _, row := range p.cache {
		if row.UpdatedAt.After(latest) {
			latest = row.UpdatedAt
		}
	}
	next := latest.Add(-1 * p.TxnWindow)
	if next.After(p.start) {
		p.start = next
		// retire outdated cache entries
		for id, row := range p.cache {
			if next.After(row.UpdatedAt) {
				delete(p.cache, id)
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
