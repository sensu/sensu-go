package ringv2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/backend/etcd"
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

	// keySeqKey is the key that points to the current sequence number of the
	// ring. The sequence number is used to compute new keys within the ring
	// prefix.
	keySeqKey string

	// intervalKey is the key that the TTL for ring items is stored at
	intervalKey string

	// triggerPrefix is the prefix that triggers are stored under. triggers are
	// leased keys that are used to notify the ring clients about which item is
	// next.
	triggerPrefix string
}

// New creates a new Ring.
func New(client *clientv3.Client, storePath string) *Ring {
	return &Ring{
		client:        client,
		itemPrefix:    path.Join(storePath, "items"),
		keySeqKey:     path.Join(storePath, "seq"),
		intervalKey:   path.Join(storePath, "interval"),
		triggerPrefix: path.Join(storePath, "triggers"),
	}
}

// dump is for debugging
func (r *Ring) dump(ctx context.Context, w io.Writer) {
	resp, err := r.client.Get(ctx, path.Dir(r.itemPrefix), clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}
	for _, kv := range resp.Kvs {
		fmt.Fprintf(w, "%s: %s (%x)\n", string(kv.Key), string(kv.Value), kv.Value)
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
	interval, err := r.getInterval(ctx)
	if err != nil {
		return nil, err
	}
	lease, err := r.client.Grant(ctx, interval)
	return lease, err
}

// Add adds a new value to the ring. If the value already exists, it will not
// be disturbed.
func (r *Ring) Add(ctx context.Context, value string) error {
	itemKey := path.Join(r.itemPrefix, value)

	getresp, err := r.client.Get(ctx, itemKey)
	if err != nil {
		return fmt.Errorf("couldn't add %q to ring: %s", value, err)
	}

	if len(getresp.Kvs) > 0 && len(getresp.Kvs[0].Value) > 0 {
		// Item already exists
		return nil
	}

	seq, err := etcd.Sequence(r.client, r.keySeqKey)
	if err != nil {
		return fmt.Errorf("couldn't add %q to ring: %s", value, err)
	}

	cmps := []clientv3.Cmp{clientv3.Compare(clientv3.Version(itemKey), "=", 0)}
	ops := []clientv3.Op{clientv3.OpPut(itemKey, seq)}
	var lease *clientv3.LeaseGrantResponse
	if empty, err := r.IsEmpty(ctx); err != nil {
		return fmt.Errorf("couldn't add %q to ring: %s", value, err)
	} else if empty {
		lease, err = r.grant(ctx)
		if err != nil {
			return fmt.Errorf("couldn't add %q to ring: %s", value, err)
		}
		triggerKey := path.Join(r.triggerPrefix, value)
		ops = append(ops, clientv3.OpPut(triggerKey, "", clientv3.WithLease(lease.ID)))
		cmps = append(cmps, clientv3.Compare(clientv3.Version(r.triggerPrefix), "=", 0).WithPrefix())
	}

	resp, err := r.client.Txn(ctx).If(cmps...).Then(ops...).Commit()

	if err != nil {
		return fmt.Errorf("couldn't add %q to ring: %s", value, err)
	}

	if !resp.Succeeded && lease != nil {
		// The item was concurrently added by another client, get rid of this
		// lease.
		_, _ = r.client.Revoke(ctx, lease.ID)
	}

	return nil
}

// Remove removes a value from the list. If the value does not exist, nothing
// happens.
func (r *Ring) Remove(ctx context.Context, value string) error {
	itemKey := path.Join(r.itemPrefix, value)
	itemCmp := clientv3.Compare(clientv3.Version(itemKey), ">", 0)
	itemOp := clientv3.OpDelete(itemKey)

	_, err := r.client.Txn(ctx).If(itemCmp).Then(itemOp).Commit()
	if err != nil {
		return fmt.Errorf("couldn't delete %q from ring: %s", value, err)
	}

	// Determine if the item we're removing is next up to be triggered
	triggerKey := path.Join(r.triggerPrefix, value)
	triggerCmp := clientv3.Compare(clientv3.Version(triggerKey), ">", 0)
	triggerOp := clientv3.OpDelete(triggerKey)

	resp, err := r.client.Txn(ctx).If(triggerCmp).Then(triggerOp).Commit()
	if err != nil {
		return fmt.Errorf("couldn't delete %q from ring: %s", value, err)
	}

	if resp.Succeeded {
		// The item was going to be the next ring item, so advance the ring
		if _, err := r.advanceRing(ctx, nil, 1); err != nil {
			return err
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
	return c
}

func (r *Ring) getInterval(ctx context.Context) (int64, error) {
	resp, err := r.client.Get(ctx, r.intervalKey)
	if err != nil {
		return 0, err
	}
	if len(resp.Kvs) == 0 {
		return 0, fmt.Errorf("ring: nil interval value at %s", r.intervalKey)
	}
	var result int64
	if _, err := fmt.Sscanf(string(resp.Kvs[0].Value), "%d", &result); err != nil {
		return 0, fmt.Errorf("ring: bad interval value at %s", r.intervalKey)
	}
	return result, nil
}

// SetInterval sets the interval between trigger events. It returns an error if
// the interval is less than MinInterval, or if there was an error from etcd.
func (r *Ring) SetInterval(ctx context.Context, seconds int64) error {
	if seconds < MinInterval {
		return fmt.Errorf("bad interval: got %ds, minimum value is %ds", seconds, MinInterval)
	}
	value := fmt.Sprintf("%d", seconds)
	_, err := r.client.Put(ctx, r.intervalKey, value)
	if err != nil {
		return fmt.Errorf("error setting TTL: %s", err)
	}
	return nil
}

func (r *Ring) hasTrigger(ctx context.Context) (bool, error) {
	resp, err := r.client.Get(ctx, r.triggerPrefix, clientv3.WithPrefix(), clientv3.WithLimit(1))
	if err != nil {
		return false, err
	}
	return len(resp.Kvs) > 0, nil
}

func (r *Ring) ensureActiveTrigger(ctx context.Context) error {
	if empty, err := r.IsEmpty(ctx); err != nil {
		return err
	} else if empty {
		return nil
	}
	if has, err := r.hasTrigger(ctx); err != nil {
		return err
	} else if !has {
		_, err := r.advanceRing(ctx, nil, 1)
		return err
	}
	return nil
}

func (r *Ring) startWatchers(ctx context.Context, ch chan<- Event, values int) {
	itemsC := r.client.Watch(ctx, r.itemPrefix, clientv3.WithPrefix())
	nextC := r.client.Watch(ctx, r.triggerPrefix, clientv3.WithPrefix(), clientv3.WithFilterPut())
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

func nextInRingOps(n int) []clientv3.OpOption {
	return []clientv3.OpOption{
		clientv3.WithPrefix(),
		clientv3.WithLimit(int64(n)),
		clientv3.WithSort(clientv3.SortByValue, clientv3.SortAscend),
	}
}

func (r *Ring) nextInRing(ctx context.Context, n int) ([]*mvccpb.KeyValue, error) {
	resp, err := r.client.Get(ctx, r.itemPrefix, nextInRingOps(n)...)
	if err != nil {
		return nil, fmt.Errorf("couldn't get next item in ring: %s", err)
	}
	return resp.Kvs, nil
}

func (r *Ring) bumpSequence(ctx context.Context, value string) error {
	seq, err := etcd.Sequence(r.client, r.keySeqKey)
	if err != nil {
		return fmt.Errorf("couldn't advance ring: %s", err)
	}
	key := path.Join(r.itemPrefix, value)
	if _, err := r.client.Put(ctx, key, seq); err != nil {
		return fmt.Errorf("couldn't advance ring: %s", err)
	}
	return nil
}

func getSequenceOps(items []*mvccpb.KeyValue, seqs []string) ([]clientv3.Op, []clientv3.Cmp) {
	if len(items) != len(seqs) {
		panic("sanity check failed")
	}
	var ops []clientv3.Op
	var cmps []clientv3.Cmp
	for i, kv := range items {
		key := string(kv.Key)
		cmps = append(cmps, clientv3.Compare(clientv3.Version(key), "=", kv.Version))
		ops = append(ops, clientv3.OpPut(key, seqs[i]))
	}
	return ops, cmps
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

	items, err := r.nextInRing(ctx, numValues+1)
	if err != nil {
		return nil, fmt.Errorf("couldn't advance ring: %s", err)
	}

	fmt.Println("len items", len(items))

	if len(items) == 0 {
		// The ring is empty
		return nil, nil
	}

	if len(items) == 1 {
		// Ring with only one item, need to repeat the item for the next round
		items = append(items, items...)
	}

	totalItems := len(items)
	repeatItems := repeatKVs(items, numValues)
	truncItems := repeatItems[len(repeatItems)-numValues:]

	if prevKv != nil && path.Base(string(prevKv.Key)) != path.Base(string(items[0].Key)) {
		// Another client has already advanced the ring, there is no need to
		// continue.
		return repeatItems, nil
	}

	seqs, err := etcd.Sequences(ctx, r.client, r.keySeqKey, numValues)
	if err != nil {
		return nil, fmt.Errorf("couldn't advance ring: %s", err)
	}

	if len(seqs) > totalItems {
		seqs = seqs[len(seqs)-totalItems:]
	}

	if len(seqs) != totalItems-1 {
		fmt.Println(len(seqs), totalItems-1)
		panic("len(seqs) != totalItems-1")
	}

	fmt.Println("len truncItems", len(truncItems))
	sequenceOps, sequenceCmps := getSequenceOps(truncItems, seqs)

	nextValue := path.Base(string(items[len(items)-1].Key))
	triggerKey := path.Join(r.triggerPrefix, nextValue)
	triggerOp := clientv3.OpPut(triggerKey, "", clientv3.WithLease(lease.ID))
	triggerCmp := clientv3.Compare(clientv3.Version(triggerKey), "=", 0)

	ops := append(sequenceOps, triggerOp)
	cmps := append(sequenceCmps, triggerCmp)

	resp, err := r.client.Txn(ctx).If(cmps...).Then(ops...).Commit()
	if err != nil {
		return nil, fmt.Errorf("couldn't advance ring: %s", err)
	}

	// Captured by the deferred function
	txnSuccess = resp.Succeeded

	return repeatItems, nil
}

func (r *Ring) handleRingTrigger(ctx context.Context, ch chan<- Event, response clientv3.WatchResponse, values int) {
	for _, event := range response.Events {
		items, err := r.advanceRing(ctx, event.Kv, values)
		if err != nil {
			notifyError(ch, err)
		}
		notifyTrigger(ch, items)
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
	ch <- Event{
		Type:   EventTrigger,
		Values: values,
	}
}

// notifyError sends EventError events to the channel
func notifyError(ch chan<- Event, err error) {
	ch <- Event{Err: err, Type: EventError}
}
