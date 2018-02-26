package queue

import (
	"bytes"
	"context"
	"encoding/binary"
	"path"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/backend/store"
)

const (
	queuePrefix    = "queue"
	workPrefix     = "work"
	inFlightPrefix = "inflight"
	itemTimeout    = 60 * time.Second
)

var (
	queueKeyBuilder = store.NewKeyBuilder(queuePrefix)
)

// Interface defines the methods available on a Queue.
type Interface interface {
	Enqueue(ctx context.Context, value string) error
	Dequeue(ctx context.Context) (*Item, error)
}

// Get interface provides access to a queue.
type Get interface {
	NewQueue(name string) Interface
}

// EtcdGetter provides access to the etcd client for creating a new queue.
type EtcdGetter struct {
	Client *clientv3.Client
}

// NewQueue provides a new queue.
func (e EtcdGetter) NewQueue(name string) Interface {
	return New(name, e.Client)
}

// Queue is a durable FIFO queue that is backed by etcd.
// When an item is received by a client, it is deleted from
// the work lane, and added to the in-flight lane. The item stays in-flight
// until it is acked by the client, or returned to the work queue in case the
// client nacks it or times out.
type Queue struct {
	client      *clientv3.Client
	work        string
	inFlight    string
	kv          clientv3.KV
	itemTimeout time.Duration
}

// New returns an instance of Queue.
func New(name string, client *clientv3.Client) *Queue {
	queue := &Queue{
		client:      client,
		work:        queueKeyBuilder.Build(name, workPrefix),
		inFlight:    queueKeyBuilder.Build(name, inFlightPrefix),
		kv:          clientv3.NewKV(client),
		itemTimeout: itemTimeout,
	}
	return queue
}

// Item contains the key and value for a dequeued item, as well as the queue it
// belongs to.
type Item struct {
	Key       string
	Value     string
	Revision  int64
	Timestamp int64
	queue     *Queue
	once      sync.Once
	mu        *sync.Mutex
	cancel    context.CancelFunc
}

// Ack acknowledges the item has been received and processed, and deletes it
// from the in flight queue.
func (i *Item) Ack(ctx context.Context) error {
	var err error
	i.once.Do(func() {
		i.mu.Lock()
		delCmp := clientv3.Compare(clientv3.ModRevision(i.Key), "=", i.Revision)
		delReq := clientv3.OpDelete(i.Value)
		_, err = i.queue.kv.Txn(ctx).If(delCmp).Then(delReq).Commit()
		i.mu.Unlock()
		i.cancel()
	})
	return err
}

// Nack returns the item to the work queue and deletes it from the in-flight
// queue.
func (i *Item) Nack(ctx context.Context) error {
	var err error
	i.once.Do(func() {
		i.mu.Lock()
		err = i.queue.swapLane(ctx, i.Key, i.Revision, i.Value, i.queue.work)
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
				uName, err := i.queue.uniqueName()
				if err != nil {
					logger.WithError(err).Error("error creating unique name for item keepalive")
				}
				updateKey := path.Join(i.queue.inFlight, uName)

				i.mu.Lock()
				// create new key, delete old key
				putCmp := clientv3.Compare(clientv3.ModRevision(updateKey), "=", 0)
				delCmp := clientv3.Compare(clientv3.ModRevision(i.Key), "=", i.Revision)
				putReq := clientv3.OpPut(updateKey, i.Value)
				delReq := clientv3.OpDelete(i.Key)

				_, err = i.queue.kv.Txn(ctx).If(putCmp, delCmp).Then(putReq, delReq).Commit()

				if err != nil {
					// log error
					logger.WithError(err).Error("error updating item keepalive timestamp")
				}

				i.Key = updateKey
				i.mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (q *Queue) swapLane(ctx context.Context, currentKey string, currentRevision int64, value string, lane string) error {
	uName, err := q.uniqueName()
	if err != nil {
		return err
	}
	uKey := path.Join(lane, uName)

	for {
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
	return q.tryPut(ctx, value)
}

func (q *Queue) tryPut(ctx context.Context, value string) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		un, err := q.uniqueName()
		if err != nil {
			return err
		}

		key := path.Join(q.work, un)
		cmp := clientv3.Compare(clientv3.Version(key), "=", 0)
		req := clientv3.OpPut(key, value)
		response, err := q.kv.Txn(ctx).If(cmp).Then(req).Commit()
		if err == nil && response.Succeeded {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (q *Queue) uniqueName() (string, error) {
	now := time.Now().UnixNano()
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, now); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Dequeue gets a value from the queue. It returns an error if the context
// is cancelled, the deadline exceeded, or if the client encounters an error.
func (q *Queue) Dequeue(ctx context.Context) (*Item, error) {
	err := q.nackExpiredItems(ctx, q.itemTimeout)
	if err != nil {
		return nil, err
	}

	response, err := q.client.Get(ctx, q.work, clientv3.WithFirstKey()...)
	if err != nil {
		return nil, err
	}
	if len(response.Kvs) > 0 {
		item, err := q.tryDelete(ctx, response.Kvs[0])
		if err != nil {
			return nil, err
		}
		if item != nil {
			return item, nil
		}
	}
	if response.More {
		// Need to retry, we are promised that there will be more.
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
	inFlightItems, err := q.client.Get(ctx, q.inFlight, clientv3.WithPrefix())
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

			err = q.swapLane(ctx, string(item.Key), item.ModRevision, string(item.Value), q.work)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (q *Queue) tryDelete(ctx context.Context, kv *mvccpb.KeyValue) (*Item, error) {
	key := string(kv.Key)

	// generate a new key name
	uName, err := q.uniqueName()
	if err != nil {
		return nil, err
	}
	uKey := path.Join(q.inFlight, uName)

	delCmp := clientv3.Compare(clientv3.ModRevision(key), "=", kv.ModRevision)
	putCmp := clientv3.Compare(clientv3.ModRevision(uKey), "=", 0)
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
		item := &Item{
			Key:       string(uKey),
			Value:     string(kv.Value),
			Revision:  revision,
			Timestamp: time.Now().UnixNano(),
			queue:     q,
			mu:        &sync.Mutex{},
			cancel:    cancel,
		}
		item.keepalive(context)
		return item, nil
	}
	return nil, nil
}

func (q *Queue) waitPutEvent(ctx context.Context) (*clientv3.Event, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	wc := q.client.Watch(ctx, q.work, clientv3.WithPrefix())
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
