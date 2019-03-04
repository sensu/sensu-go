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
	"github.com/robfig/cron"
	"github.com/sensu/sensu-go/backend/store"
)

var _ Interface = &Ring{}

// EventType is an enum that describes the type of event received by watchers.
type EventType int

const (
	EventError EventType = iota
	EventAdd
	EventRemove
	EventTrigger
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

type Ring struct {
	client *clientv3.Client

	// itemPrefix is the prefix of all the ring items. The ring items are KVs
	// that map from the value of the item to its sequence ID. The sequence
	// IDs are updated on every ring advancement, and are used to sort the ring
	// in order to find the least-valued item.
	itemPrefix string

	// intervalKey is the key that the TTL for ring items is stored at
	intervalKey string

	// interval is the TTL for ring triggers
	interval int64

	// triggerKey is the key that contains ring triggers
	triggerKey string

	// watchCtr counts the number of open watchers
	watchCtr int64

	// cron is a cron schedule
	cron cron.Schedule

	mu sync.Mutex
}

// New creates a new Ring.
func New(client *clientv3.Client, storePath string) *Ring {
	return &Ring{
		client:      client,
		itemPrefix:  path.Join(storePath, "items"),
		intervalKey: path.Join(storePath, "interval"),
		triggerKey:  path.Join(storePath, "trigger"),
		interval:    5,
	}
}

// IsEmpty returns true if there are no items in the ring.
func (r *Ring) IsEmpty(ctx context.Context) (bool, error) {
	resp, err := r.client.Get(ctx, r.itemPrefix,
		clientv3.WithKeysOnly(),
		clientv3.WithPrefix(),
		clientv3.WithLimit(1))
	if err != nil {
		return false, err
	}
	return len(resp.Kvs) == 0, nil
}

func (r *Ring) grant(ctx context.Context) (*clientv3.LeaseGrantResponse, error) {
	interval := r.getInterval()
	lease, err := r.client.Grant(ctx, interval)
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

	getresp, err := r.client.Get(ctx, itemKey)
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

	lease, err := r.client.Grant(ctx, keepalive)
	if err != nil {
		return fmt.Errorf("couldn't add %q to ring: %s", value, err)
	}
	defer func() {
		if rerr != nil && lease != nil {
			_, _ = r.client.Revoke(ctx, lease.ID)
		}
	}()

	if _, err := r.client.Put(ctx, itemKey, "", clientv3.WithLease(lease.ID)); err != nil {
		return fmt.Errorf("couldn't add %q to ring: %s", value, err)
	}

	if atomic.LoadInt64(&r.watchCtr) > 0 {
		if err := r.ensureActiveTrigger(ctx); err != nil {
			return fmt.Errorf("couldn't add %q to ring: %s", value, err)
		}
	}

	return nil
}

// Remove removes a value from the list. If the value does not exist, nothing
// happens.
func (r *Ring) Remove(ctx context.Context, value string) error {
	itemKey := path.Join(r.itemPrefix, value)
	itemOp := clientv3.OpDelete(itemKey)

	triggerCmp := clientv3.Compare(clientv3.Value(r.triggerKey), "=", value)
	triggerOp := clientv3.OpDelete(r.triggerKey)

	resp, err := r.client.Txn(ctx).If(triggerCmp).Then(itemOp, triggerOp).Else(itemOp).Commit()
	if err != nil {
		return fmt.Errorf("couldn't delete %q from ring: %s", value, err)
	}

	if resp.Succeeded && atomic.LoadInt64(&r.watchCtr) > 0 {
		if err := r.ensureActiveTrigger(ctx); err != nil {
			return fmt.Errorf("fatal ring error: %s", err)
		}
	}

	return nil
}

// Watch watches the ring for events. The events are sent on the channel that
// is returned.
//
// If the underlying etcd watcher fails, then the Event will contain a non-nil
// error.
//
// If the context is canceled, EventClosing will be sent on the channel, and it
// will be closed.
//
// The values parameter controls how many ring values the event will contain.
// If the requested number of values is greater than the number of items in
// the values will contain repetitions in order to satisfy the request.
func (r *Ring) Watch(ctx context.Context, values int) <-chan Event {
	c := make(chan Event, 1)
	r.startWatchers(ctx, c, values)
	atomic.AddInt64(&r.watchCtr, 1)
	return c
}

func (r *Ring) getInterval() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cron != nil {
		now := time.Now()
		interval := int64(r.cron.Next(now).Sub(now) / time.Second)
		if interval < MinInterval {
			interval = int64(r.cron.Next(now.Add((time.Duration(interval)*time.Second)+1)).Sub(now) / time.Second)
		}
		return interval
	}
	return r.interval
}

// SetInterval sets the interval between trigger events. It returns an error if
// the interval is less than MinInterval, or if there was an error from etcd.
func (r *Ring) SetInterval(ctx context.Context, seconds int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if seconds < MinInterval {
		return fmt.Errorf("bad interval: got %ds, minimum value is %ds", seconds, MinInterval)
	}
	r.cron = nil
	r.interval = seconds
	return nil
}

// SetCron sets a cron schedule instead of an interval.
func (r *Ring) SetCron(schedule cron.Schedule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cron = schedule
}

// hasTrigger returns whether the ring has an active trigger. If the ring
// does not have an active trigger, the first lexical key in the ring will
// be returned in the string return value.
func (r *Ring) hasTrigger(ctx context.Context) (bool, string, error) {
	getTrigger := clientv3.OpGet(r.triggerKey)
	getFirst := clientv3.OpGet(r.itemPrefix, clientv3.WithPrefix(), clientv3.WithLimit(1))
	resp, err := r.client.Txn(ctx).Then(getTrigger, getFirst).Commit()
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

func (r *Ring) ensureActiveTrigger(ctx context.Context) error {
	has, next, err := r.hasTrigger(ctx)
	if err != nil {
		return err
	}
	if has || next == "" {
		// if next == "", there are no ring items
		return nil
	}
	lease, err := r.grant(ctx)
	if err != nil {
		return err
	}
	nextValue := path.Base(next)
	triggerOp := clientv3.OpPut(r.triggerKey, nextValue, clientv3.WithLease(lease.ID))
	triggerCmp := clientv3.Compare(clientv3.Version(r.triggerKey), "=", 0)

	resp, err := r.client.Txn(ctx).If(triggerCmp).Then(triggerOp).Commit()
	if !resp.Succeeded {
		_, _ = r.client.Revoke(ctx, lease.ID)
	}
	return err
}

func (r *Ring) startWatchers(ctx context.Context, ch chan<- Event, values int) {
	itemsC := r.client.Watch(ctx, r.itemPrefix, clientv3.WithPrefix())
	nextC := r.client.Watch(ctx, r.triggerKey, clientv3.WithFilterPut(), clientv3.WithPrevKV())
	if err := r.ensureActiveTrigger(ctx); err != nil {
		notifyError(ch, fmt.Errorf("error while starting ring watcher: %s", err))
		notifyClosing(ch)
		return
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				notifyClosing(ch)
				close(ch)
				atomic.AddInt64(&r.watchCtr, -1)
				return
			case response := <-itemsC:
				if err := response.Err(); err != nil {
					notifyError(ch, err)
					continue
				}
				notifyAddRemove(ch, response)
			case response := <-nextC:
				if err := response.Err(); err != nil {
					notifyError(ch, err)
					continue
				}
				r.handleRingTrigger(ctx, ch, response, values)
			}
		}
	}()
}

func notifyClosing(ch chan<- Event) {
	ch <- Event{Type: EventClosing}
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
	resp, err := r.client.Get(ctx, key, opts...)
	if err != nil {
		return nil, fmt.Errorf("couldn't get next item(s) in ring: %s", err)
	}
	result := resp.Kvs
	if len(result) == 0 {
		return nil, nil
	}
	if int64(len(result)) < n {
		m := n - int64(len(result))
		resp, err := r.client.Get(ctx, r.itemPrefix, clientv3.WithPrefix(), clientv3.WithLimit(m))
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

func (r *Ring) advanceRing(ctx context.Context, prevKv *mvccpb.KeyValue, numValues int) ([]*mvccpb.KeyValue, error) {
	lease, err := r.grant(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't advance ring: %s", err)
	}

	txnSuccess := false
	defer func() {
		if !txnSuccess {
			_, _ = r.client.Revoke(ctx, lease.ID)
		}
	}()

	items, err := r.nextInRing(ctx, prevKv, int64(numValues)+1)
	if err != nil {
		return nil, fmt.Errorf("couldn't advance ring: %s", err)
	}

	if len(items) == 0 {
		// The ring is empty
		return nil, nil
	}

	nextItem := items[len(items)-1]
	repeatItems := repeatKVs(items, numValues)
	if len(items) < numValues+1 {
		// There are fewer items than requested values
		nextItem = items[0]
	} else {
		items = items[:len(items)-1]
	}

	nextValue := path.Base(string(nextItem.Key))
	triggerOp := clientv3.OpPut(r.triggerKey, nextValue, clientv3.WithLease(lease.ID))
	triggerCmp := clientv3.Compare(clientv3.Version(r.triggerKey), "=", 0)

	resp, err := r.client.Txn(ctx).If(triggerCmp).Then(triggerOp).Commit()
	if err != nil {
		return nil, fmt.Errorf("couldn't advance ring: %s", err)
	}

	// Captured by the deferred function
	txnSuccess = resp.Succeeded

	return repeatItems, nil
}

func (r *Ring) handleRingTrigger(ctx context.Context, ch chan<- Event, response clientv3.WatchResponse, values int) {
	for _, event := range response.Events {
		items, err := r.advanceRing(ctx, event.PrevKv, values)
		if err != nil {
			notifyError(ch, err)
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
func notifyError(ch chan<- Event, err error) {
	ch <- Event{Err: err, Type: EventError}
}
