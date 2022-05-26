package postgres

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lib/pq"
	"github.com/sensu/sensu-go/backend/poll"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// watchStoreOverrides contains per-store overrides to the default watcher query builder.
var watchStoreOverridesMu sync.Mutex
var watchStoreOverrides map[string]watchStoreFactory = make(map[string]watchStoreFactory)

type watchStoreFactory func(storev2.ResourceRequest, *pgxpool.Pool) (poll.Table, error)

func registerWatchStoreOverride(storeName string, factory watchStoreFactory) {
	watchStoreOverridesMu.Lock()
	defer watchStoreOverridesMu.Unlock()
	watchStoreOverrides[storeName] = factory
}

func getWatchStoreOverride(storeName string) (factory watchStoreFactory, ok bool) {
	watchStoreOverridesMu.Lock()
	defer watchStoreOverridesMu.Unlock()
	factory, ok = watchStoreOverrides[storeName]
	return
}

func watch(ctx context.Context, poller *poll.Poller, watchChan chan []storev2.WatchEvent) {
	defer close(watchChan)
	poller.Initialize(ctx)
	for {
		changes, err := poller.Next(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			watchChan <- []storev2.WatchEvent{{Err: err}}
			return
		}
		if len(changes) == 0 {
			continue
		}
		notifications := make([]storev2.WatchEvent, len(changes))
		for i, change := range changes {
			notifications[i].Value = change.Resource.(storev2.Wrapper)
			r, err := notifications[i].Value.Unwrap()
			if err != nil {
				// shouldn't happen
				panic(err)
			}
			meta := r.GetMetadata()
			notifications[i].Key = storev2.ResourceRequest{
				Namespace: meta.Namespace,
				Name:      meta.Name,
				StoreName: r.StoreName(),
			}
			switch change.Change {
			case poll.Create:
				notifications[i].Type = storev2.Create
			case poll.Update:
				notifications[i].Type = storev2.Update
			case poll.Delete:
				notifications[i].Type = storev2.Delete
			}
		}
		watchChan <- notifications
	}
}

// recordStatus used by postgres stores implementing poll.Table
type recordStatus struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt pq.NullTime
}

// Row builds a poll.Row from a scanned row
func (rs recordStatus) Row(id string, resource storev2.Wrapper) poll.Row {
	row := poll.Row{
		Id:        id,
		Resource:  resource,
		CreatedAt: rs.CreatedAt,
		UpdatedAt: rs.UpdatedAt,
	}
	if rs.DeletedAt.Valid {
		row.DeletedAt = &rs.DeletedAt.Time
	}
	return row
}
