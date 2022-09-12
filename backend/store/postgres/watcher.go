package postgres

import (
	"context"
	"fmt"
	"time"

	v3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/poll"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/util/retry"
)

type WatchableStore interface {
	GetPoller(request storev2.ResourceRequest) (poll.Table, error)
}

type Watcher struct {
	store          WatchableStore
	watchInterval  time.Duration
	watchTxnWindow time.Duration
}

func NewWatcher(store WatchableStore, watchInterval time.Duration, watchTxnWindow time.Duration) *Watcher {
	return &Watcher{
		store:          store,
		watchInterval:  watchInterval,
		watchTxnWindow: watchTxnWindow,
	}
}

func (w *Watcher) Watch(ctx context.Context, req storev2.ResourceRequest) <-chan []storev2.WatchEvent {
	eventChan := make(chan []storev2.WatchEvent, 32)

	table, err := w.store.GetPoller(req)
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
