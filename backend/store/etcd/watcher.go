package etcd

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/sensu/sensu-go/backend/store"
	"go.etcd.io/etcd/client/v3"
	"golang.org/x/time/rate"
)

const (
	// Set a buffer for the outgoing channel in order to reduce times of context
	// switches
	resultChanBufSize = 100
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
	revision   int64
	resultChan chan store.WatchEvent
	opts       []clientv3.OpOption
	logger     *logrus.Entry
	wg         sync.WaitGroup
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
	w.start()

	return w
}

// newWatcher creates a new Watcher
func newWatcher(ctx context.Context, client *clientv3.Client, key string, recursive bool, opts ...clientv3.OpOption) *Watcher {
	wc := &Watcher{
		client:     client,
		key:        key,
		recursive:  recursive,
		resultChan: make(chan store.WatchEvent, resultChanBufSize),
		opts:       opts,
		logger:     logger.WithField("key", key),
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
	baseOpts := []clientv3.OpOption{clientv3.WithCreatedNotify(), clientv3.WithPrevKV()}
	if w.recursive {
		baseOpts = append(baseOpts, clientv3.WithPrefix())
	}
	opts := make([]clientv3.OpOption, len(baseOpts))
	copy(opts, baseOpts)
	if w.revision != 0 {
		opts = append(opts, clientv3.WithRev(w.revision))
	}

	// Create a channel to be notified if the watch channel is closed
	watchChanStopped := make(chan struct{})

	// Create a cancellable context for our watcher
	ctx, cancel := context.WithCancel(w.ctx)

	// Start the watcher
	w.logger.Debug("starting a watcher")
	w.watch(ctx, opts, watchChanStopped)

	// Initialize a rate limiter
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)

	go func() {
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
				w.logger.Warning("restarting the watcher")

				// Specify the latest revision we tracked
				opts = make([]clientv3.OpOption, len(baseOpts))
				copy(opts, baseOpts)
				if w.revision != 0 {
					opts = append(opts, clientv3.WithRev(w.revision))
				}

				// Wait for the previous goroutine to stop before starting another
				w.wg.Wait()

				w.watch(ctx, opts, watchChanStopped)
			case <-w.ctx.Done():
				// The consumer has cancelled this watcher, we need to exit
				w.logger.Debug("stopping the watcher")
				break RetryLoop
			}
		}

		// Use both parent and current watcher context's cancellable functions to reap
		// all goroutines. At this point the consumer has stopped the watcher so we
		// should stop everything. It's also fine to double cancel.
		cancel()
		w.cancel()
		w.wg.Wait()
		close(w.resultChan)
	}()
}

func (w *Watcher) watch(ctx context.Context, opts []clientv3.OpOption, watchChanStopped chan struct{}) {
	w.wg.Add(1)
	// Wrap the context with WithRequireLeader so ErrNoLeader is returned and the
	// WatchChan is closed if the etcd server has no leader
	ctx = clientv3.WithRequireLeader(ctx)

	watchChan := w.client.Watch(ctx, w.key, opts...)

	go func() {
		// Loop over the watchChan channel to receive watch responses. The loop will
		// exit if the channel is closed.
		defer w.wg.Done()
		for watchResponse := range watchChan {
			if watchResponse.Err() != nil {
				// We received an error from the channel, so let's assume it's no longer
				// functional and exit this goroutine so we can try to re-create the
				// watcher
				w.logger.WithError(watchResponse.Err()).Warn("error from watch response")

				// Check if we received a compact revision that bigger than the revision
				// we are keeping track of
				if watchResponse.CompactRevision > w.revision {
					// If we arrive at this point, it means we missed watch events.
					// Therefore we need to send a WatchError so the watcher consumer can
					// act accordingly.
					w.revision = watchResponse.CompactRevision
					w.logger.Debugf("watch revision updated to %d by compact revision", w.revision)
					w.queueEvent(ctx, store.WatchEvent{Type: store.WatchError})
				}
				break
			}

			// Bump the revision to indicate that this revision was handled by the
			// watcher. We do bump the revision in response to a WatchCreateRequest,
			// only if there's new events
			if !watchResponse.Created && watchResponse.Header.GetRevision() >= w.revision {
				w.revision = watchResponse.Header.GetRevision() + 1
				w.logger.Debugf("watch revision updated to %d by header revision", w.revision)
			}

			for _, event := range watchResponse.Events {
				w.logger.Debugf("received event of type %s", event.Type.String())
				parsedEvent := parseEvent(event)
				w.queueEvent(ctx, parsedEvent)
			}
		}

		// Verify if the parent context was cancelled, which would indicate that the
		// watcher was gracefully shutdown by its consumer and therefore we don't
		// need to print a warning here
		if err := w.ctx.Err(); err == nil || err != context.Canceled {
			w.logger.Warning("the watcher has been stopped")
		}

		// At this point, the watch channel has been closed by its consumer or is
		// broken, therefore we should notify the main thread that this goroutine has
		// exited
		close(watchChanStopped)
	}()

}

// queueEvent takes an incoming event from the watcher and adds it to the buffer
// of outgoing results
func (w *Watcher) queueEvent(ctx context.Context, e store.WatchEvent) {
	if len(w.resultChan) == cap(w.resultChan) {
		w.logger.Warning("resultChan buffer is full, watch events are not " +
			"processed fast enough, incoming events from the watcher will be blocked")
	}

	select {
	case w.resultChan <- e:
	case <-ctx.Done():
	}
}

func parseEvent(e *clientv3.Event) store.WatchEvent {
	event := store.WatchEvent{
		Key:      string(e.Kv.Key),
		Revision: e.Kv.ModRevision,
		Object:   e.Kv.Value,
	}

	if e.IsCreate() {
		event.Type = store.WatchCreate
	} else if e.IsModify() {
		event.Type = store.WatchUpdate
	} else {
		event.Type = store.WatchDelete
		// If the previous key value is not available, return a watch error
		if e.PrevKv == nil {
			return store.WatchEvent{Type: store.WatchError}
		}
		// Fetch the key value before it was deleted
		event.Object = e.PrevKv.Value
	}

	return event
}
