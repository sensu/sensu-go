package etcd

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"golang.org/x/time/rate"
)

// Watcher implements the store.Watcher interface rather than clientv3.Watcher,
// so the channel returned by the Watch method only provides a single event at a
// time instead of a list of events, and the events are ready to be consumed
type Watcher struct {
	client     *clientv3.Client
	key        string
	recursive  bool
	eventChan  chan *clientv3.Event
	resultChan chan store.WatchEvent
	opts       []clientv3.OpOption
}

// Watch returns a Watcher for the given key. If recursive is true, then the
// watcher is created with clientv3.WithPrefix. The watcher will also be provided
// with any etcd client options passed in.
func Watch(ctx context.Context, client *clientv3.Client, key string, recursive bool, opts ...clientv3.OpOption) *Watcher {
	// Make sure we have a trailing slash if we need to watch the key and its
	// children
	if recursive && !strings.HasSuffix(key, "/") {
		key += "/"
	}

	// From etcd docs:
	// If the context is "context.Background/TODO", returned "WatchChan" will
	// not be closed and block until event is triggered, except when server
	// returns a non-recoverable error (e.g. ErrCompacted).
	// For example, when context passed with "WithRequireLeader" and the
	// connected server has no leader (e.g. due to network partition),
	// error "etcdserver: no leader" (ErrNoLeader) will be returned,
	// and then "WatchChan" is closed with non-nil "Err()".
	// In order to prevent a watch stream being stuck in a partitioned node,
	// make sure to wrap context with "WithRequireLeader".
	ctx = clientv3.WithRequireLeader(ctx)

	w := newWatcher(client, key, recursive, opts...)
	w.start(ctx)

	return w
}

// newWatcher creates a new Watcher
func newWatcher(client *clientv3.Client, key string, recursive bool, opts ...clientv3.OpOption) *Watcher {
	return &Watcher{
		client:     client,
		key:        key,
		recursive:  recursive,
		eventChan:  make(chan *clientv3.Event),
		resultChan: make(chan store.WatchEvent),
		opts:       opts,
	}
}

// Result returns the resultChan
func (w *Watcher) Result() <-chan store.WatchEvent {
	return w.resultChan
}

// start starts watching the configured key and sends all etcd events
// received to resultChan
func (w *Watcher) start(ctx context.Context) {
	opts := []clientv3.OpOption{clientv3.WithCreatedNotify()}
	if w.recursive {
		opts = append(opts, clientv3.WithPrefix())
	}

	opts = append(opts, w.opts...)

	logger.Debugf("starting a watcher for key %s", w.key)

	watcherChan := w.client.Watch(ctx, w.key, opts...)
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)

	go func() {
		defer close(w.resultChan)
		_ = limiter.Wait(ctx)
		for ctx.Err() == nil {
			for watchResponse := range watcherChan {
				if err := watchResponse.Err(); err != nil {
					if ctx.Err() != nil {
						// Our context was canceled, return without error,
						// since the consumer is probably shutting down.
						return
					}
					logger.WithError(err).Info("error from watch response")
					w.resultChan <- store.WatchEvent{
						Type: store.WatchError,
						Err:  err,
					}
					if watchResponse.Canceled {
						// Reinstate the watcher and break to the outer loop
						watcherChan = w.client.Watch(ctx, w.key, opts...)
						break
					}
					continue
				}

				for _, event := range watchResponse.Events {
					logger.Debugf("received event of type %v for key %s", event.Type, event.Kv.Key)
					w.event(ctx, event)
				}
			}
			if w.client.Ctx().Err() != nil {
				w.resultChan <- store.WatchEvent{
					Type: store.WatchError,
					Err:  errors.New("client closed unexpectedly"),
				}
				return
			}
		}
	}()
}

func (w *Watcher) event(ctx context.Context, e *clientv3.Event) {
	typ := GetWatcherAction(e)
	if typ == store.WatchUnknown {
		logger.Infof("unknown etcd watch action type %q", e.Type.String())
		return
	}

	result := store.WatchEvent{
		Type:   typ,
		Key:    string(e.Kv.Key),
		Object: e.Kv.Value,
	}

	select {
	case w.resultChan <- result:
	case <-ctx.Done():
		return
	}
}
