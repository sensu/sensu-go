package queue

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

type QueueClient interface {
	clientv3.KV
	clientv3.Cluster
	clientv3.Watcher
}

const (
	queuePrefix     = "queue"
	workPostfix     = "work"
	inFlightPostfix = "inflight"
	itemTimeout     = 60 * time.Second
)

var (
	queueKeyBuilder = store.NewKeyBuilder(queuePrefix)
)

// EtcdGetter provides access to the etcd client for creating a new queue.
type EtcdGetter struct {
	Client    *clientv3.Client
	BackendID string
}

// GetQueue gets a new Queue.
func (e EtcdGetter) GetQueue(path ...string) types.Queue {
	return New(queueKeyBuilder.Build(path...), e.Client, e.BackendID)
}

// Queue is a durable FIFO queue that is backed by etcd.
// When an item is received by a client, it is deleted from
// the work lane, and added to the in-flight lane. The item stays in-flight
// until it is acked by the client, or returned to the work queue in case the
// client nacks it or times out.
type Queue struct {
	kv          clientv3.KV
	cluster     clientv3.Cluster
	watcher     clientv3.Watcher
	itemTimeout time.Duration
	name        string
	backendID   string
}

func (q *Queue) workPrefix() string {
	return path.Join(q.name, q.backendID, workPostfix)
}

func (q *Queue) inFlightPrefix() string {
	return path.Join(q.name, q.backendID, inFlightPostfix)
}

// New returns an instance of Queue.
func New(name string, client QueueClient, backendID string) *Queue {
	queue := &Queue{
		name:        name,
		kv:          client,
		cluster:     client,
		watcher:     client,
		itemTimeout: itemTimeout,
		backendID:   backendID,
	}
	return queue
}

// Item is a Queue item.
type Item struct {
	key       string
	value     string
	revision  int64
	timestamp int64
	queue     *Queue
	once      sync.Once
	mu        *sync.Mutex
	cancel    context.CancelFunc
}

// Value returns the value of the Item.
func (i *Item) Value() string {
	return i.value
}

// Ack acknowledges the Item has been received and processed, and deletes it
// from the in flight queue.
func (i *Item) Ack(ctx context.Context) error {
	var err error
	i.once.Do(func() {
		i.mu.Lock()
		delCmp := clientv3.Compare(clientv3.ModRevision(i.key), "=", i.revision)
		delReq := clientv3.OpDelete(i.key)
		_, err = i.queue.kv.Txn(ctx).If(delCmp).Then(delReq).Commit()
		i.mu.Unlock()
		i.cancel()
	})
	return err
}

// Nack returns the Item to the work queue and deletes it from the in-flight
// queue.
func (i *Item) Nack(ctx context.Context) error {
	var err error
	i.once.Do(func() {
		i.mu.Lock()
		err = i.queue.swapLane(ctx, i.key, i.revision, i.value, i.queue.workPrefix())
		i.mu.Unlock()
		i.cancel()
	})
	return err
}

func (i *Item) keepalive(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	// stop the goroutine when the context is canceled (if Ack or Nack is
	// called)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// create new key with new timestamp
				seq, err := i.queue.timeStamp()
				if err != nil {
					logger.WithError(err).Error("error creating unique name for item keepalive")
					return
				}
				updateKey := path.Join(i.queue.inFlightPrefix(), seq)

				i.mu.Lock()
				// create new key, delete old key
				putCmp := clientv3.Compare(clientv3.ModRevision(updateKey), "=", 0)
				delCmp := clientv3.Compare(clientv3.ModRevision(i.key), "=", i.revision)
				putReq := clientv3.OpPut(updateKey, i.value)
				delReq := clientv3.OpDelete(i.key)

				_, err = i.queue.kv.Txn(ctx).If(putCmp, delCmp).Then(putReq, delReq).Commit()

				if err != nil {
					// log error
					if err != context.Canceled {
						logger.WithError(err).Error("error updating item keepalive timestamp")
					}
					i.mu.Unlock()
					return
				}

				i.key = updateKey
				i.mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()
}

// swapLane swaps a key/value pair from one place to another
func (q *Queue) swapLane(ctx context.Context, currentKey string, currentRevision int64, value string, lane string) error {
	for {
		seq, err := q.timeStamp()
		if err != nil {
			return fmt.Errorf("queue error: %s", err)
		}
		uKey := path.Join(lane, seq)

		putCmp := clientv3.Compare(clientv3.ModRevision(uKey), "=", 0)
		delCmp := clientv3.Compare(clientv3.ModRevision(currentKey), "=", currentRevision)
		putReq := clientv3.OpPut(uKey, value)
		delReq := clientv3.OpDelete(currentKey)

		response, err := q.kv.Txn(ctx).If(putCmp, delCmp).Then(putReq, delReq).Commit()
		if response.Succeeded {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Enqueue adds a new value to the queue. It returns an error if the context is
// canceled, the deadline exceeded, or if the client encounters an error.
func (q *Queue) Enqueue(ctx context.Context, value string) error {
	resp, err := q.cluster.MemberList(ctx)
	if err != nil {
		return fmt.Errorf("queue: couldn't enqueue item: %s", err)
	}
	if err := q.cleanup(ctx, resp.Members); err != nil {
		return fmt.Errorf("queue: couldn't enqueue item: %s", err)
	}
	cmps, ops, err := q.enqueueOps(resp.Members, value)
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

// cleanup stale queue entries belonging to backends that no longer exist
func (q *Queue) cleanup(ctx context.Context, members []*etcdserverpb.Member) error {
	memberSet := make(map[string]struct{})
	for _, member := range members {
		memberSet[fmt.Sprintf("%x", member.ID)] = struct{}{}
	}
	resp, err := q.kv.Get(ctx, q.name, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	ops := []clientv3.Op{}
	for _, kv := range resp.Kvs {
		// Look for backends that are not present anymore in the member list
		key := string(kv.Key)
		parts := strings.Split(key, "/")
		backendID := parts[len(parts)-3]
		if _, ok := memberSet[backendID]; !ok {
			ops = append(ops, clientv3.OpDelete(key))
		}
	}
	_, err = q.kv.Txn(ctx).If().Then(ops...).Commit()
	return err
}

func (q *Queue) enqueueOps(members []*etcdserverpb.Member, value string) ([]clientv3.Cmp, []clientv3.Op, error) {
	cmps := []clientv3.Cmp{}
	ops := []clientv3.Op{}

	for _, m := range members {
		backendID := fmt.Sprintf("%x", m.ID)
		seq, err := q.timeStamp()
		if err != nil {
			return nil, nil, fmt.Errorf("queue error: %s", err)
		}
		workKey := path.Join(q.name, backendID, workPostfix, seq)
		cmp := clientv3.Compare(clientv3.ModRevision(workKey), "=", 0)
		op := clientv3.OpPut(workKey, value)
		cmps = append(cmps, cmp)
		ops = append(ops, op)
	}

	return cmps, ops, nil
}

// Dequeue gets a value from the queue. It returns an error if the context
// is cancelled, the deadline exceeded, or if the client encounters an error.
func (q *Queue) Dequeue(ctx context.Context) (types.QueueItem, error) {
	err := q.nackExpiredItems(ctx, q.itemTimeout)
	if err != nil {
		return nil, err
	}

	response, err := q.kv.Get(ctx, q.workPrefix(), clientv3.WithFirstKey()...)
	if err != nil {
		return nil, err
	}

	for _, kv := range response.Kvs {
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
	binaryTimestamp := key[len(key)-8:]

	var itemTimestamp int64
	buf := bytes.NewReader(binaryTimestamp)
	err := binary.Read(buf, binary.BigEndian, &itemTimestamp)
	if err != nil {
		return time.Time{}, err
	}
	unixTimestamp := time.Unix(0, itemTimestamp)
	return unixTimestamp, nil
}

func (q *Queue) nackExpiredItems(ctx context.Context, timeout time.Duration) error {
	// get all items in the inflight queue
	inFlightItems, err := q.kv.Get(ctx, q.inFlightPrefix(), clientv3.WithPrefix())
	if err != nil {
		return err
	}

	// get the timestamp from each key
	for _, item := range inFlightItems.Kvs {
		itemTimestamp, err := q.getItemTimestamp(item.Key)
		if err != nil {
			return err
		}
		// If the item has timed out or the client has disconnected, the item is
		// considered expired and should be moved back to the work queue.
		if time.Since(itemTimestamp) > timeout || ctx.Err() != nil {
			err = q.swapLane(ctx, string(item.Key), item.ModRevision, string(item.Value), q.workPrefix())
			if err != nil {
				return err
			}
		}
	}
	return nil
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
	putReq := clientv3.OpPut(uKey, string(kv.Value))
	delReq := clientv3.OpDelete(key)

	response, err := q.kv.Txn(ctx).If(putCmp, delCmp).Then(putReq, delReq).Commit()
	if err != nil {
		return nil, err
	}

	// return the new item
	if response.Succeeded {
		putResp := response.Responses[0].GetResponsePut()
		revision := putResp.GetHeader().GetRevision()
		context, cancel := context.WithCancel(ctx)
		value := string(kv.Value)
		item := &Item{
			key:       string(uKey),
			value:     value,
			revision:  revision,
			timestamp: time.Now().UnixNano(),
			queue:     q,
			mu:        &sync.Mutex{},
			cancel:    cancel,
		}
		item.keepalive(context)
		return item, nil
	}
	return nil, nil
}

// The queue uses timestamps to order its queue items, and also to
// determine how old queue items are.
func (q *Queue) timeStamp() (string, error) {
	now := time.Now().UnixNano()
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, now); err != nil {
		return "", err
	}
	return buf.String(), nil
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
