package ringv2

import (
	"context"
	"errors"
	"fmt"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	cron "github.com/robfig/cron/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/util/retry"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var logger = logrus.WithFields(logrus.Fields{
	"component": "ring",
})

// EventType is an enum that describes the type of event received by watchers.
type EventType int

const (
	// EventError is a message sent when a ring processing error occurs.
	EventError EventType = iota
	// EventAdd is a message sent when an item is added to the ring.
	EventAdd
	// EventRemove is a message sent when an item is removed from the ring.
	EventRemove
	// EventTrigger is a message sent when a ring item has moved from the front of the queue to the back.
	EventTrigger
	// EventClosing is a message sent when the ring is closing due to context cancellation.
	EventClosing
)

func (e EventType) String() string {
	switch e {
	case EventAdd:
		return "EventAdd"
	case EventRemove:
		return "EventRemove"
	case EventTrigger:
		return "EventTrigger"
	case EventError:
		return "EventError"
	case EventClosing:
		return "EventClosing"
	default:
		return "INVALID"
	}
}

// MinInterval ...
const MinInterval = 5

// Path returns the canonical path to a ring.
func Path(namespace, subscription string) string {
	return store.NewKeyBuilder("rings").WithNamespace(namespace).Build(subscription)
}

// Event represents an event that occurred in a ring. The event can originate
// from any ring client.
type Event struct {
	// Type is the type of the event.
	Type EventType

	// Values are the ring items associated with the event. For trigger events,
	// the length of Values will be equal to the results per interval.
	Values []string

	// Err is any error that occurred while processing the event.
	Err error
}

// Ring is a circular queue of items that are cooperatively iterated over by one
// or more subscribers. The subscribers are notified every time an item goes from
// the front to the back of the queue, with EventTrigger. Iteration proceeds
// according to an interval, and is triggered by an etcd lease expiring.
type Ring struct {
	client *clientv3.Client

	// itemPrefix is the prefix of all the ring items. The ring items are KVs
	// that map from the value of the item to its sequence ID. The sequence
	// IDs are updated on every ring advancement, and are used to sort the ring
	// in order to find the least-valued item.
	itemPrefix string

	// interval is the TTL for ring triggers
	interval int

	// triggerPrefix is the prefix that contains ring triggers
	triggerPrefix string

	// watchCtr counts the number of open watchers
	watchCtr int64

	// cron is a cron schedule
	cron cron.Schedule

	// watchers is the set of active watchers
	watchers map[watcherKey]*watcher

	mu sync.Mutex

	// limit watch restarts to one per second (defensive)
	watchLimiter *rate.Limiter
}

type watcherKey struct {
	name     string
	values   int
	interval int
	cron     string
}

func (w *watcher) triggerKey() string {
	interval := w.watcherKey.cron
	if interval == "" {
		interval = fmt.Sprintf("%d", w.interval)
	}
	return path.Join(w.ring.triggerPrefix, w.name, fmt.Sprintf("%d", w.values), interval)
}

type watcher struct {
	watcherKey
	notifier chan struct{}
	ring     *Ring
	cron     cron.Schedule
	events   <-chan Event
}

func newWatcher(ring *Ring, ch <-chan Event, name string, values, interval int, schedule string) (*watcher, error) {
	var sched cron.Schedule
	if schedule != "" {
		var err error
		sched, err = cron.ParseStandard(schedule)
		if err != nil {
			return nil, err
		}
	}
	return &watcher{
		watcherKey: watcherKey{
			name:     name,
			values:   values,
			cron:     schedule,
			interval: interval,
		},
		notifier: make(chan struct{}, 10),
		ring:     ring,
		cron:     sched,
		events:   ch,
	}, nil
}

// New creates a new Ring.
func New(client *clientv3.Client, storePath string) *Ring {
	return &Ring{
		client:        client,
		itemPrefix:    path.Join(storePath, "items"),
		triggerPrefix: path.Join(storePath, "triggers"),
		interval:      5,
		watchers:      make(map[watcherKey]*watcher),
		watchLimiter:  rate.NewLimiter(rate.Every(time.Second), 1),
	}
}

// IsEmpty returns true if there are no items in the ring.
func (r *Ring) IsEmpty(ctx context.Context) (bool, error) {
	var resp *clientv3.GetResponse
	err := etcd.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = r.client.Get(ctx, r.itemPrefix,
			clientv3.WithKeysOnly(),
			clientv3.WithPrefix(),
			clientv3.WithLimit(1))
		return etcd.RetryRequest(n, err)
	})
	if err != nil {
		return false, err
	}
	return len(resp.Kvs) == 0, nil
}

func (w *watcher) grant(ctx context.Context) (*clientv3.LeaseGrantResponse, error) {
	interval := w.getInterval()
	lease, err := w.ring.client.Grant(ctx, int64(interval))
	return lease, err
}

// Add adds a new value to the ring. If the value already exists, its keepalive
// will be reset. Values that are not kept alive will expire and be removed
// from the ring.
func (r *Ring) Add(ctx context.Context, value string, keepalive int64) (rerr error) {
	if keepalive < 5 {
		return fmt.Errorf("couldn't add %q to ring: keepalive must be >5s", value)
	}

	itemKey := path.Join(r.itemPrefix, value)

	var getresp *clientv3.GetResponse
	err := etcd.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		getresp, err = r.client.Get(ctx, itemKey)
		return etcd.RetryRequest(n, err)
	})
	if err != nil {
		return fmt.Errorf("couldn't add %q to ring: %s", value, err)
	}

	if len(getresp.Kvs) > 0 && getresp.Kvs[0].Version > 0 {
		leaseID := clientv3.LeaseID(getresp.Kvs[0].Lease)
		if leaseID == 0 {
			goto NEWLEASE
		}
		// Item already exists
		resp, err := r.client.KeepAliveOnce(ctx, leaseID)
		if err != nil {
			// error most likely due to lease not existing
			goto NEWLEASE
		}
		if resp.TTL == keepalive {
			// We can return early since the TTL is as requested.
			return nil
		}
		// The TTL is different than the requested keepalive, so revoke the
		// lease and create a new one afterwards.
		_, _ = r.client.Revoke(ctx, leaseID)
	}
NEWLEASE:

	var lease *clientv3.LeaseGrantResponse
	err = etcd.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		lease, err = r.client.Grant(ctx, keepalive)
		return etcd.RetryRequest(n, err)
	})
	if err != nil {
		return fmt.Errorf("couldn't add %q to ring: %s", value, err)
	}
	defer func() {
		if rerr != nil && lease != nil {
			_, _ = r.client.Revoke(ctx, lease.ID)
		}
	}()

	err = etcd.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		_, err = r.client.Put(ctx, itemKey, "", clientv3.WithLease(lease.ID))
		return etcd.RetryRequest(n, err)
	})
	if err != nil {
		return fmt.Errorf("couldn't add %q to ring: %s", value, err)
	}

	r.notifyWatchers()

	return nil
}

func (r *Ring) notifyWatchers() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, watcher := range r.watchers {
		watcher.notifier <- struct{}{}
	}
}

// Remove removes a value from the list. If the value does not exist, nothing
// happens.
func (r *Ring) Remove(ctx context.Context, value string) error {
	return etcd.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		_, err = r.client.Delete(ctx, path.Join(r.itemPrefix, value))
		return etcd.RetryRequest(n, err)
	})
}

// Watch watches the ring for events. The events are sent on the channel that
// is returned. For each interval duration in seconds, one or more values will
// be delivered, if there are any values in the ring.
//
// If the underlying etcd watcher fails, then the Event will contain a non-nil
// error.
//
// If the context is canceled, EventClosing will be sent on the channel, and it
// will be closed.
//
// The name parameter specifies the name of the watch. Watchers should use
// unique names when requesting different numbers of values.
//
// The interval parameter sets the interval at which ring items will be
// delivered, in seconds.
//
// If the cron parameter is not the empty string, the interval parameter will
// be ignored, and the watcher will deliver values according to the cron
// schedule.
//
// The values parameter controls how many ring values the event will contain.
// If the requested number of values is greater than the number of items in
// the values will contain repetitions in order to satisfy the request.
func (r *Ring) Watch(ctx context.Context, name string, values, interval int, cron string) <-chan Event {
	key := watcherKey{name: name, values: values, interval: interval, cron: cron}
	r.mu.Lock()
	w, ok := r.watchers[key]
	r.mu.Unlock()
	if ok {
		return w.events
	}
	c := make(chan Event, 1)
	r.startWatchers(ctx, c, name, values, interval, cron)
	atomic.AddInt64(&r.watchCtr, 1)
	return c
}

func (w *watcher) getInterval() int {
	if w.cron != nil {
		now := time.Now()
		// Add 1s to the interval to deal with the effects of truncation
		interval := int(w.cron.Next(now).Sub(now)/time.Second) + 1
		for interval < MinInterval {
			now = now.Add(time.Second)
			interval = int(w.cron.Next(now).Sub(now)/time.Second) + 1
		}
		return interval
	}
	return w.interval
}

// hasTrigger returns whether the ring has an active trigger. If the ring
// does not have an active trigger, the first lexical key in the ring will
// be returned in the string return value.
func (w *watcher) hasTrigger(ctx context.Context) (bool, string, error) {
	getTrigger := clientv3.OpGet(w.triggerKey())
	getFirst := clientv3.OpGet(w.ring.itemPrefix, clientv3.WithPrefix(), clientv3.WithLimit(1))
	var resp *clientv3.TxnResponse
	err := etcd.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = w.ring.client.Txn(ctx).Then(getTrigger, getFirst).Commit()
		return etcd.RetryRequest(n, err)
	})
	if err != nil {
		return false, "", err
	}
	if got := len(resp.Responses); got != 2 {
		return false, "", fmt.Errorf("bad response length: got %d, want 2", got)
	}
	triggerResp := resp.Responses[0].GetResponseRange()
	valueResp := resp.Responses[1].GetResponseRange()
	if len(triggerResp.Kvs) == 0 || len(triggerResp.Kvs[0].Value) == 0 {
		value := ""
		if len(valueResp.Kvs) > 0 {
			value = string(valueResp.Kvs[0].Key)
		}
		return false, value, nil
	}
	return true, "", nil
}

func (w *watcher) ensureActiveTrigger(ctx context.Context) error {
	backoff := retry.ExponentialBackoff{
		InitialDelayInterval: 10 * time.Millisecond,
		MaxDelayInterval:     10 * time.Second,
		Multiplier:           10,
		Ctx:                  ctx,
	}

	err := backoff.Retry(func(retry int) (bool, error) {
		has, next, err := w.hasTrigger(ctx)
		if err != nil {
			if err == context.Canceled {
				return true, err
			}
			logger.WithError(err).Error("can't check ring trigger, retrying")
			return false, nil
		}
		if has || next == "" {
			// if next == "", there are no ring items
			return true, nil
		}
		lease, err := w.grant(ctx)
		if err != nil {
			logger.WithError(err).Error("can't grant ring trigger lease, retrying")
			return false, nil
		}
		nextValue := path.Base(next)
		triggerOp := clientv3.OpPut(w.triggerKey(), nextValue, clientv3.WithLease(lease.ID))
		triggerCmp := clientv3.Compare(clientv3.Version(w.triggerKey()), "=", 0)

		resp, err := w.ring.client.Txn(ctx).If(triggerCmp).Then(triggerOp).Commit()
		if err != nil {
			return etcd.RetryRequest(retry, err)
		}
		if !resp.Succeeded {
			_, _ = w.ring.client.Revoke(ctx, lease.ID)
		}
		return true, nil
	})

	return err
}

func (r *Ring) startWatchers(ctx context.Context, ch chan Event, name string, values, interval int, cron string) {
	_ = r.watchLimiter.Wait(ctx)
	watcher, err := newWatcher(r, ch, name, values, interval, cron)
	if err != nil {
		notifyError(ctx, ch, err)
		notifyClosing(ctx, ch)
		return
	}
	cancelCtx, cancel := context.WithCancel(clientv3.WithRequireLeader(ctx))
	itemsC := r.client.Watch(cancelCtx, r.itemPrefix, clientv3.WithPrefix())
	nextC := r.client.Watch(cancelCtx, watcher.triggerKey(), clientv3.WithFilterPut(), clientv3.WithPrevKV())
	r.mu.Lock()
	r.watchers[watcher.watcherKey] = watcher
	r.mu.Unlock()
	if err := watcher.ensureActiveTrigger(ctx); err != nil {
		notifyError(ctx, ch, fmt.Errorf("error while starting ring watcher: %s", err))
		notifyClosing(ctx, ch)
		cancel()
		return
	}

	go func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				r.mu.Lock()
				delete(r.watchers, watcher.watcherKey)
				r.mu.Unlock()
				notifyClosing(ctx, ch)
				return
			case response, ok := <-itemsC:
				if err := response.Err(); err != nil {
					notifyError(ctx, ch, err)
				}
				if response.Canceled || !ok {
					// The watcher needs to be reinstated
					r.startWatchers(ctx, ch, name, values, interval, cron)
					return
				}
				notifyAddRemove(ch, response)
			case response, ok := <-nextC:
				if err := response.Err(); err != nil {
					notifyError(ctx, ch, err)
				}
				if response.Canceled || !ok {
					// The watcher needs to be reinstated
					r.startWatchers(ctx, ch, name, values, interval, cron)
					return
				}
				watcher.handleRingTrigger(ctx, ch, response)
			case <-watcher.notifier:
				if err := watcher.ensureActiveTrigger(ctx); err != nil {
					notifyError(ctx, ch, err)
				}
			}
		}
	}()
}

func notifyClosing(ctx context.Context, ch chan<- Event) {
	select {
	case ch <- Event{Type: EventClosing}:
	case <-ctx.Done():
	}
	close(ch)
}

func (r *Ring) nextInRing(ctx context.Context, prevKv *mvccpb.KeyValue, n int64) ([]*mvccpb.KeyValue, error) {
	opts := []clientv3.OpOption{clientv3.WithLimit(n)}
	var key string
	if prevKv == nil {
		key = r.itemPrefix
		opts = append(opts, clientv3.WithPrefix())
	} else {
		value := string(prevKv.Value)
		key = path.Join(r.itemPrefix, value)
		end := path.Join(r.itemPrefix, string([]byte{0xFF}))
		opts = append(opts, clientv3.WithFromKey())
		opts = append(opts, clientv3.WithRange(end))
	}
	var resp *clientv3.GetResponse
	err := etcd.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = r.client.Get(ctx, key, opts...)
		return etcd.RetryRequest(n, err)
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't get next item(s) in ring: %s", err)
	}
	result := resp.Kvs
	if len(result) == 0 && prevKv == nil {
		return nil, nil
	} else if len(result) == 0 {
		// If a delete occurred and it corresponded to the trigger, need to try
		// again with nil prevKv
		return r.nextInRing(ctx, nil, n)
	}
	if int64(len(result)) < n {
		m := n - int64(len(result))
		var resp *clientv3.GetResponse
		err := etcd.Backoff(ctx).Retry(func(n int) (done bool, err error) {
			resp, err = r.client.Get(ctx, r.itemPrefix, clientv3.WithPrefix(), clientv3.WithLimit(m))
			return etcd.RetryRequest(n, err)
		})
		if err != nil {
			return nil, fmt.Errorf("couldn't get next item(s) in ring: %s", err)
		}
		result = append(result, resp.Kvs...)
	}
	return result, nil
}

func repeatKVs(kvs []*mvccpb.KeyValue, items int) []*mvccpb.KeyValue {
	result := make([]*mvccpb.KeyValue, 0, items)
	for i := 0; i < (items / len(kvs)); i++ {
		result = append(result, kvs...)
	}
	for i := 0; i < (items % len(kvs)); i++ {
		result = append(result, kvs[i])
	}
	return result
}

func (w *watcher) advanceRing(ctx context.Context, prevKv *mvccpb.KeyValue) ([]*mvccpb.KeyValue, error) {
	items, err := w.ring.nextInRing(ctx, prevKv, int64(w.values)+1)
	if err != nil {
		return nil, fmt.Errorf("couldn't advance ring: %s", err)
	}

	if len(items) == 0 {
		// The ring is empty
		return nil, nil
	}

	nextItem := items[len(items)-1]
	repeatItems := repeatKVs(items, w.values)
	if len(items) < w.values+1 {
		// There are fewer items than requested values
		nextItem = items[0]
	}

	lease, err := w.grant(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't advance ring: %s", err)
	}

	txnSuccess := false
	defer func() {
		if !txnSuccess {
			_, _ = w.ring.client.Revoke(ctx, lease.ID)
		}
	}()

	nextValue := path.Base(string(nextItem.Key))
	triggerOp := clientv3.OpPut(w.triggerKey(), nextValue, clientv3.WithLease(lease.ID))
	triggerCmp := clientv3.Compare(clientv3.Version(w.triggerKey()), "=", 0)

	var resp *clientv3.TxnResponse
	err = etcd.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = w.ring.client.Txn(ctx).If(triggerCmp).Then(triggerOp).Commit()
		return etcd.RetryRequest(n, err)
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't advance ring: %s", err)
	}

	// Captured by the deferred function
	txnSuccess = resp.Succeeded

	return repeatItems, nil
}

func (w *watcher) handleRingTrigger(ctx context.Context, ch chan<- Event, response clientv3.WatchResponse) {
	for _, event := range response.Events {
		items, err := w.advanceRing(ctx, event.PrevKv)
		if err != nil {
			notifyError(ctx, ch, err)
		}
		if len(items) > 0 {
			// When the ring trigger was deleted by the Remove() method, the
			// items will be empty.
			notifyTrigger(ch, items)
		}
	}
}

// notifyAddRemove sends EventAdd or EventRemove events to the channel
func notifyAddRemove(ch chan<- Event, response clientv3.WatchResponse) {
	for _, event := range response.Events {
		if event.Kv == nil {
			ch <- Event{
				Err: errors.New("nil Kv from ring watcher"),
			}
			continue
		}
		if event.Kv.Version > 1 {
			// The item was put, and already existed
			continue
		}
		eventType := EventRemove
		if event.Type == mvccpb.PUT {
			eventType = EventAdd
		}
		ch <- Event{
			Type:   eventType,
			Values: []string{path.Base(string(event.Kv.Key))},
		}
	}
}

// notifyTrigger sents EventTrigger events to the channel
func notifyTrigger(ch chan<- Event, items []*mvccpb.KeyValue) {
	if len(items) == 0 {
		ch <- Event{
			Err: errors.New("trigger without items"),
		}
		return
	}
	values := make([]string, len(items))
	for i := range values {
		values[i] = path.Base(string(items[i].Key))
	}
	event := Event{
		Type:   EventTrigger,
		Values: values,
	}
	ch <- event
}

// notifyError sends EventError events to the channel
func notifyError(ctx context.Context, ch chan<- Event, err error) {
	select {
	case ch <- Event{Err: err, Type: EventError}:
	case <-ctx.Done():
	}
}
