package queue

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"path"
	"strconv"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	queuePrefix     = "queue"
	workPostfix     = "work"
	inFlightPostfix = "inflight"
	itemTimeout     = 60 * time.Second
)

var (
	queueKeyBuilder    = store.NewKeyBuilder(queuePrefix)
	backendIDKeyPrefix = store.NewKeyBuilder("backends").Build()
)

type BackendIDGetter interface {
	GetBackendID() int64
}

// EtcdGetter provides access to the etcd client for creating a new queue.
type EtcdGetter struct {
	Client          *clientv3.Client
	BackendIDGetter BackendIDGetter
}

// GetQueue gets a new Queue.
func (e EtcdGetter) GetQueue(path ...string) types.Queue {
	return New(queueKeyBuilder.Build(path...), e.Client, e.BackendIDGetter)
}

// Queue is a non-durable FIFO queue that is backed by etcd.
// When an item is received by a client, it is deleted from
// the work lane, and added to the in-flight lane. The item stays in the
// in-flight lane until it is Acked by the client, or returned to the work
// lane with Nack.
type Queue struct {
	kv              clientv3.KV
	lease           clientv3.Lease
	watcher         clientv3.Watcher
	itemTimeout     time.Duration
	name            string
	backendIDGetter BackendIDGetter
}

func (q *Queue) backendID() int64 {
	return q.backendIDGetter.GetBackendID()
}

func (q *Queue) workPrefix() string {
	return path.Join(q.name, fmt.Sprintf("%x", q.backendID()), workPostfix)
}

func (q *Queue) inFlightPrefix() string {
	return path.Join(q.name, fmt.Sprintf("%x", q.backendID()), inFlightPostfix)
}

// New returns an instance of Queue.
func New(name string, client *clientv3.Client, backendIDGetter BackendIDGetter) *Queue {
	queue := &Queue{
		name:            name,
		kv:              client,
		lease:           client,
		watcher:         client,
		itemTimeout:     itemTimeout,
		backendIDGetter: backendIDGetter,
	}
	return queue
}

// Item is a Queue item.
type Item struct {
	key   string
	value string
	queue *Queue
}

// Key returns the key of the Item.
func (i *Item) Key() string {
	return i.key
}

// Value returns the value of the Item.
func (i *Item) Value() string {
	return i.value
}

// Ack acknowledges the Item has been received and processed, and deletes it
// from the in flight lane.
func (i *Item) Ack(ctx context.Context) error {
	_, err := i.queue.kv.Delete(ctx, i.key)
	return err
}

// Nack returns the Item to the work queue and deletes it from the in-flight
// lane.
func (i *Item) Nack(ctx context.Context) error {
	return i.queue.swapLane(ctx, i.key, i.value, i.queue.workPrefix())
}

// swapLane swaps a key/value pair from one lane to another
func (q *Queue) swapLane(ctx context.Context, currentKey, value string, lane string) error {
	for {
		seq, err := q.timeStamp()
		if err != nil {
			return fmt.Errorf("queue error: %s", err)
		}
		uKey := path.Join(lane, seq)

		putCmp := clientv3.Compare(clientv3.ModRevision(uKey), "=", 0)
		leaseID := clientv3.LeaseID(q.backendID())
		putReq := clientv3.OpPut(uKey, value, clientv3.WithLease(leaseID))
		delReq := clientv3.OpDelete(currentKey)

		response, err := q.kv.Txn(ctx).If(putCmp).Then(putReq, delReq).Commit()
		if response.Succeeded {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// getAllBackendIDs gets all currently valid BackendIDs.
func getAllBackendIDs(ctx context.Context, kv clientv3.KV) ([]string, error) {
	resp, err := kv.Get(ctx, backendIDKeyPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		result = append(result, string(kv.Value))
	}
	return result, nil
}

// Enqueue adds a new value to the queue. It returns an error if the context is
// canceled, the deadline exceeded, or if the client encounters an error.
func (q *Queue) Enqueue(ctx context.Context, value string) error {
	backendIDs, err := getAllBackendIDs(ctx, q.kv)
	if err != nil {
		return fmt.Errorf("queue: couldn't enqueue item: %s", err)
	}
	cmps, ops, err := q.enqueueOps(backendIDs, value)
	if err != nil {
		return fmt.Errorf("queue: couldn't enqueue item: %s", err)
	}
	for {
		if ctx.Err() != nil {
			return fmt.Errorf("queue: couldn't enqueue item: %s", ctx.Err())
		}
		response, err := q.kv.Txn(ctx).If(cmps...).Then(ops...).Commit()
		if err == nil && response.Succeeded {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (q *Queue) enqueueOps(backendIDs []string, value string) ([]clientv3.Cmp, []clientv3.Op, error) {
	cmps := []clientv3.Cmp{}
	ops := []clientv3.Op{}

	for _, backendID := range backendIDs {
		seq, err := q.timeStamp()
		if err != nil {
			return nil, nil, fmt.Errorf("queue error: %s", err)
		}
		leaseID, err := strconv.ParseInt(backendID, 16, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("queue: invalid backend ID %q: %s", backendID, err)
		}
		workKey := path.Join(q.name, backendID, workPostfix, seq)
		cmp := clientv3.Compare(clientv3.ModRevision(workKey), "=", 0)
		op := clientv3.OpPut(workKey, value, clientv3.WithLease(clientv3.LeaseID(leaseID)))
		cmps = append(cmps, cmp)
		ops = append(ops, op)
	}

	return cmps, ops, nil
}

// Dequeue gets a value from the queue. It returns an error if the context
// is cancelled, the deadline exceeded, or if the client encounters an error.
func (q *Queue) Dequeue(ctx context.Context) (types.QueueItem, error) {
	response, err := q.kv.Get(ctx, q.workPrefix(), clientv3.WithFirstKey()...)
	if err != nil {
		return nil, err
	}

	if len(response.Kvs) > 0 {
		kv := response.Kvs[0]
		item, err := q.tryDelete(ctx, kv)
		if err != nil {
			return nil, err
		}
		if item != nil {
			return item, nil
		}
		// no item - we lost the race to another consumer
		return q.Dequeue(ctx)
	}

	// Wait for the queue to receive an item
	event, err := q.waitPutEvent(ctx)
	if err != nil {
		return nil, err
	}

	if event != nil {
		item, err := q.tryDelete(ctx, event.Kv)
		if err != nil {
			return nil, err
		}
		return item, nil
	}
	return q.Dequeue(ctx)
}

func (q *Queue) getItemTimestamp(key []byte) (time.Time, error) {
	ts, err := hex.DecodeString(string(key[len(key)-16:]))
	if err != nil {
		return time.Time{}, err
	}

	var itemTimestamp int64
	buf := bytes.NewReader(ts)
	if err := binary.Read(buf, binary.BigEndian, &itemTimestamp); err != nil {
		return time.Time{}, err
	}
	unixTimestamp := time.Unix(0, itemTimestamp)
	return unixTimestamp, nil
}

func (q *Queue) tryDelete(ctx context.Context, kv *mvccpb.KeyValue) (types.QueueItem, error) {
	key := string(kv.Key)

	// generate a new key name
	seq, err := q.timeStamp()
	if err != nil {
		return nil, fmt.Errorf("error deleting queue item: %s", err)
	}
	uKey := path.Join(q.inFlightPrefix(), seq)

	delCmp := clientv3.Compare(clientv3.Version(key), ">", 0)
	putCmp := clientv3.Compare(clientv3.Version(uKey), "=", 0)
	leaseID := clientv3.LeaseID(q.backendID())
	putReq := clientv3.OpPut(uKey, string(kv.Value), clientv3.WithLease(leaseID))
	delReq := clientv3.OpDelete(key)

	response, err := q.kv.Txn(ctx).If(putCmp, delCmp).Then(putReq, delReq).Commit()
	if err != nil {
		return nil, err
	}

	// return the new item
	if response.Succeeded {
		value := string(kv.Value)
		item := &Item{
			key:   string(uKey),
			value: value,
			queue: q,
		}
		return item, nil
	}
	return nil, nil
}

// The queue uses timestamps to order its queue items, and also to
// determine how old queue items are.
func (q *Queue) timeStamp() (string, error) {
	now := uint64(time.Now().UnixNano())
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, now); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf.Bytes()), nil
}

func (q *Queue) waitPutEvent(ctx context.Context) (*clientv3.Event, error) {
	wc := q.watcher.Watch(ctx, q.workPrefix(), clientv3.WithPrefix())
	// wc is a channel
	if wc == nil {
		return nil, ctx.Err()
	}
	for response := range wc {
		events := response.Events
		for _, event := range events {
			if event.Type == mvccpb.PUT {
				return event, nil
			}
		}
	}
	return nil, ctx.Err()
}
