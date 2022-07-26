package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	v3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/poll"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/util/retry"
)

// watchStoreOverrides contains per-store overrides to the default watcher query builder.
var watchStoreOverridesMu sync.Mutex
var watchStoreOverrides = make(map[string]watchStoreFactory)

var watchConfigStoreOverride watchStoreFactory

type watchStoreFactory func(storev2.ResourceRequest, *pgxpool.Pool) (poll.Table, error)

type WatchedStore interface {
	GetPgxPool() *pgxpool.Pool
}

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

func registerWatchConfigStoreOverride(factory watchStoreFactory) {
	watchConfigStoreOverride = factory
}

// recordStatus used by postgres stores implementing poll.Table
type recordStatus struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
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

type Watcher struct {
	store          WatchedStore
	watchInterval  time.Duration
	watchTxnWindow time.Duration
}

func NewWatcher(store WatchedStore, watchInterval time.Duration, watchTxnWindow time.Duration) *Watcher {
	return &Watcher{
		store:          store,
		watchInterval:  watchInterval,
		watchTxnWindow: watchTxnWindow,
	}
}

func (w *Watcher) WatchConfig(ctx context.Context, req storev2.ResourceRequest) <-chan []storev2.WatchEvent {
	return w.watch(ctx, req, true)
}

func (w *Watcher) Watch(ctx context.Context, req storev2.ResourceRequest) <-chan []storev2.WatchEvent {
	return w.watch(ctx, req, false)
}

func (w *Watcher) watch(ctx context.Context, req storev2.ResourceRequest, isConfig bool) <-chan []storev2.WatchEvent {
	eventChan := make(chan []storev2.WatchEvent, 32)

	var table poll.Table

	tableFactory, ok := getWatchStoreOverride(req.StoreName)
	if !ok {
		if isConfig {
			tableFactory = watchConfigStoreOverride
		} else {
			tableFactory = func(storev2.ResourceRequest, *pgxpool.Pool) (poll.Table, error) {
				return nil, fmt.Errorf("default watcher not yet implemented")
			}
		}
	}
	table, err := tableFactory(req, w.store.GetPgxPool())
	if err != nil {
		panic(fmt.Errorf("could not create watcher for request %v: %v", req, err))
	}

	interval, txnWindow := w.watchInterval, w.watchTxnWindow
	if interval <= 0 {
		interval = time.Second
	}
	if txnWindow <= 0 {
		txnWindow = 5 * time.Second
	}
	poller := &poll.Poller{
		Interval:  interval,
		TxnWindow: txnWindow,
		Table:     table,
	}

	backoff := retry.ExponentialBackoff{
		Ctx: ctx,
	}
	err = backoff.Retry(func(retry int) (bool, error) {
		if err := poller.Initialize(ctx); err != nil {
			logger.Errorf("watcher initialize polling error on retry %d: %v", retry, err)
			return false, err
		}
		return true, nil
	})
	if err != nil {
		logger.Errorf("watcher failed to start: %v", err)
		close(eventChan)
		return eventChan
	}

	go w.watchLoop(ctx, req, poller, eventChan)
	return eventChan
}

func (w *Watcher) watchLoop(ctx context.Context, req storev2.ResourceRequest, poller *poll.Poller, watchChan chan []storev2.WatchEvent) {
	defer close(watchChan)
	for {
		changes, err := poller.Next(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Error(err)
		}
		if len(changes) == 0 {
			continue
		}
		notifications := make([]storev2.WatchEvent, len(changes))
		for i, change := range changes {
			wrapper, ok := change.Resource.(storev2.Wrapper)
			if !ok {
				// Poller table must return Resource of type Wrapper.
				panic("postgres store watcher resource is not storev2.Wrapper")
			}
			notifications[i].Value = wrapper
			r, err := notifications[i].Value.Unwrap()
			if err != nil {
				notifications[i].Err = err
			}
			if r != nil {
				meta := r.GetMetadata()
				typeMeta := v3.V2ResourceProxy{Resource: r}.GetTypeMeta()
				notifications[i].Key = storev2.ResourceRequest{
					Namespace:  meta.Namespace,
					Name:       meta.Name,
					StoreName:  r.StoreName(),
					APIVersion: typeMeta.APIVersion,
					Type:       typeMeta.Type,
				}
			}
			switch change.Change {
			case poll.Create:
				notifications[i].Type = storev2.WatchCreate
			case poll.Update:
				notifications[i].Type = storev2.WatchUpdate
			case poll.Delete:
				notifications[i].Type = storev2.WatchDelete
			}
		}
		var status string
		select {
		case watchChan <- notifications:
			status = storev2.WatchEventsStatusHandled
		default:
			status = storev2.WatchEventsStatusDropped
		}
		storev2.WatchEventsProcessed.WithLabelValues(
			status,
			req.StoreName,
			req.Namespace,
			storev2.WatcherProviderPG,
		).Add(float64(len(notifications)))
	}
}
