package etcd

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
)

// watcher implements the store.Watcher interface rather than clientv3.Watcher,
// so the channel returned by the Watch method only provides a single event at a
// time instead of a list of events, and the events are ready to be consumed
type watcher struct {
	client     *clientv3.Client
	cancel     context.CancelFunc
	ctx        context.Context
	key        string
	recursive  bool
	errChan    chan error
	eventChan  chan *clientv3.Event
	resultChan chan store.WatchEvent
}

// Watch returns a watcher for the given key
func Watch(ctx context.Context, client *clientv3.Client, key string, recursive bool) store.Watcher {
	// Make sure we have a trailing slash if we need to watch the key and its
	// children
	if recursive && !strings.HasSuffix(key, "/") {
		key += "/"
	}

	w := createWatcher(ctx, client, key, recursive)

	// Start a waitgroup to ensure the watcher is properly started
	var startedWG sync.WaitGroup
	startedWG.Add(1)

	go w.run(&startedWG)

	// Make sure the watcher was properly started before returning it
	startedWG.Wait()
	return w
}

// createWatcher creates and initializes a watcher
func createWatcher(ctx context.Context, client *clientv3.Client, key string, recursive bool) *watcher {
	w := &watcher{
		client:     client,
		key:        key,
		recursive:  recursive,
		errChan:    make(chan error, 1),
		eventChan:  make(chan *clientv3.Event),
		resultChan: make(chan store.WatchEvent),
	}
	w.ctx, w.cancel = context.WithCancel(ctx)

	return w
}

// Result returns the resultChan
func (w *watcher) Result() <-chan store.WatchEvent {
	return w.resultChan
}

// Stop ends all goroutines
func (w *watcher) Stop() {
	// Stop all goroutines
	w.cancel()
}

// run starts a watcher and handles the result and errors coming over the
// various channels
func (w *watcher) run(wg *sync.WaitGroup) {
	go w.startWatching(wg)

	// Start a waitgroup to ensure the resultChan is not closed while being used
	var resultChanWG sync.WaitGroup
	resultChanWG.Add(1)
	go w.processEvent(&resultChanWG)

	select {
	case err := <-w.errChan:
		if err == context.Canceled {
			break
		}

		errResult := store.WatchEvent{
			Type:   store.WatchError,
			Object: []byte(err.Error()),
		}
		select {
		case w.resultChan <- errResult:
		case <-w.ctx.Done(): // client has given up all results
		}
	case <-w.ctx.Done(): // client cancel
	}

	// Stop all goroutines
	w.cancel()

	// Make sure that resultChan is no longer used before closing it
	resultChanWG.Wait()
	close(w.resultChan)
}

// startWatching starts watching the given key and sends all etcd events
// received to eventChan
func (w *watcher) startWatching(wg *sync.WaitGroup) {
	opts := []clientv3.OpOption{clientv3.WithCreatedNotify()}
	if w.recursive {
		opts = append(opts, clientv3.WithPrefix())
	}

	logger.Debugf("starting a watcher for key %s", w.key)

	watcher := clientv3.NewWatcher(w.client)
	watcherChan := watcher.Watch(w.ctx, w.key, opts...)

	// Inform the process the watcher was properly started
	wg.Done()

	for watchResponse := range watcherChan {
		if watchResponse.Err() != nil {
			w.errChan <- watchResponse.Err()
			return
		}

		for _, event := range watchResponse.Events {
			logger.Debugf("received event of type %v for key %s", event.Type, event.Kv.Key)
			w.event(event)
		}
	}

	// When we reach this point, the etcd watcher chan is broken and no errors
	// were sent via the channel and this watcher wasn't stopped, so we need to
	// notify the main thread so other goroutine are exited too
	w.errChan <- errors.New("etcd watcher channel was unexpectedly closed")
}

// processEvent handles incoming etcd event transmitted via eventChan,
// transforms them into a store.WatchEvent struct and then send them back via
// resultChan
func (w *watcher) processEvent(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case e := <-w.eventChan:
			typ := GetWatcherAction(e)
			if typ == store.WatchUnknown {
				logger.Infof("unknown etcd watch action type %q", e.Type.String())
				continue
			}

			result := store.WatchEvent{
				Type:   typ,
				Key:    string(e.Kv.Key),
				Object: e.Kv.Value,
			}

			select {
			case w.resultChan <- result:
			case <-w.ctx.Done():
				return
			}
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *watcher) event(e *clientv3.Event) {
	select {
	case w.eventChan <- e:
	case <-w.ctx.Done():
	}
}
