package etcd

import (
	"context"
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
	ctx        context.Context
	cancel     context.CancelFunc
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

	w := newWatcher(ctx, client, key, recursive, opts...)
	go w.start()

	return w
}

// newWatcher creates a new Watcher
func newWatcher(ctx context.Context, client *clientv3.Client, key string, recursive bool, opts ...clientv3.OpOption) *Watcher {
	wc := &Watcher{
		client:     client,
		key:        key,
		recursive:  recursive,
		eventChan:  make(chan *clientv3.Event),
		resultChan: make(chan store.WatchEvent),
		opts:       opts,
	}
	wc.ctx, wc.cancel = context.WithCancel(ctx)
	return wc
}

// Result returns the resultChan
func (w *Watcher) Result() <-chan store.WatchEvent {
	return w.resultChan
}

func (w *Watcher) start() {
	// Define the client options for this watcher
	opts := []clientv3.OpOption{clientv3.WithCreatedNotify()}
	if w.recursive {
		opts = append(opts, clientv3.WithPrefix())
	}
	opts = append(opts, w.opts...)

	// Create a channel to be notified if the watch channel is closed
	watchChanStopped := make(chan struct{})

	// Create a cancellable context for our watcher
	ctx, cancel := context.WithCancel(w.ctx)

	// Start the watcher
	logger.Debugf("starting a watcher for key %s", w.key)
	go w.watch(ctx, opts, watchChanStopped)

	// Initialize a rate limiter
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)

RetryLoop:
	for {
		select {
		case <-watchChanStopped:
			// The watch channel is broken, so let's make sure to close the watcher
			// first by cancelling its context, to prevent any possible memory leak
			cancel()

			// Now, we don't want to reconnect too quickly so wait for a moment
			_ = limiter.Wait(w.ctx)

			// Re-create a cancellable context for our watcher
			ctx, cancel = context.WithCancel(w.ctx)

			// Re-create a channel to be notified if the watch channel is closed
			watchChanStopped = make(chan struct{})

			// Restart the watcher
			logger.Debugf("restarting the watcher for key %s", w.key)
			go w.watch(ctx, opts, watchChanStopped)
		case <-w.ctx.Done():
			// The consumer has cancelled this watcher, we need to exit
			logger.Debugf("stopping the watcher for key %s", w.key)
			break RetryLoop
		}
	}

	// Use both parent and current watcher context's cancellable functions to reap
	// all goroutines. At this point the consumer has stopped the watcher so we
	// should stop everything. It's also fine to double cancel.
	cancel()
	w.cancel()
	close(w.resultChan)
}

func (w *Watcher) watch(ctx context.Context, opts []clientv3.OpOption, watchChanStopped chan struct{}) {
	// Wrap the context with WithRequireLeader so ErrNoLeader is returned and the
	// WatchChan is closed if the etcd server has no leader
	ctx = clientv3.WithRequireLeader(ctx)

	watchChan := w.client.Watch(ctx, w.key, opts...)

	// Loop over the watchChan channel to receive watch responses. The loop will
	// exit if the channel is closed.
	for watchResponse := range watchChan {
		if watchResponse.Err() != nil {
			// We received an error from the channel, so let's assume it's no longer
			// functional and exit this goroutine so we can try to re-create the
			// watcher
			logger.WithError(watchResponse.Err()).Info("error from watch response")
			break
		}

		for _, event := range watchResponse.Events {
			logger.Debugf("received event of type %v for key %s", event.Type, event.Kv.Key)
			w.event(ctx, event)
		}
	}

	// At this point, the watch channel has been closed by its consumer or is
	// broken, therefore we should notify the main thread that this goroutine has
	// exited
	close(watchChanStopped)
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
