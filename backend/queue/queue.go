package queue

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"path"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/backend/store"
)

const (
	queuePrefix    = "queue"
	workPrefix     = "work"
	inFlightPrefix = "inflight"
)

var (
	queueKeyBuilder = store.NewKeyBuilder(queuePrefix)
)

// Queue is a FIFO queue that is backed by etcd.
// Queue is not durable. When an item is received by a client, it is deleted
// from etcd. Clients are responsible for handling the item, and there is no
// way to retrieve an item again once it has been Dequeued.
type Queue struct {
	client   *clientv3.Client
	Work     string
	InFlight string
	kv       clientv3.KV
}

// New returns an instance of Queue.
func New(name string, client *clientv3.Client) *Queue {
	queue := &Queue{
		client:   client,
		Work:     queueKeyBuilder.Build(name, workPrefix),
		InFlight: queueKeyBuilder.Build(name, inFlightPrefix),
		kv:       clientv3.NewKV(client),
	}
	return queue
}

// Interface provides methods for acting on an item retrieved from the queue.
type Interface interface {
	Ack(context.Context) error
	Nack(context.Context) error
	Value() string
}

// Item contains the key and data for a dequeued item
type Item struct {
	data *mvccpb.KeyValue
}

// Ack acknowledges the item has been received and processed, and deletes it
// from the queue.
func (i *Item) Ack(ctx context.Context) error {
	return nil
}

// Nack returns the item to the queue.
func (i *Item) Nack(ctx context.Context) error {
	// how should this refer to the queue?
	// put value in work queue
	tryPut(ctx, i.Value())
	// remove data from the in flight queue using its key
	client.OpDelete(ctx, i.data.Key)
	return nil
}

// Value returns the string value of the item
func (i *Item) Value() string {
	fmt.Println(i.data)
	fmt.Println(string(i.data.Value))
	return string(i.data.Value)
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

		key := path.Join(q.Work, un)

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
	response, err := q.client.Get(ctx, q.Work, clientv3.WithFirstKey()...)
	if err != nil {
		return &Item{}, err
	}
	if len(response.Kvs) > 0 {
		item, err := q.tryDelete(ctx, response.Kvs[0])
		if err != nil {
			return &Item{}, err
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
		return &Item{}, err
	}

	if event != nil {
		item, err := q.tryDelete(ctx, event.Kv)
		fmt.Println(item)
		if err != nil {
			return &Item{}, err
		}
		return item, nil
	}

	return q.Dequeue(ctx)
}

func (q *Queue) tryDelete(ctx context.Context, kv *mvccpb.KeyValue) (*Item, error) {
	key := string(kv.Key)
	cmp := clientv3.Compare(clientv3.ModRevision(key), "=", kv.ModRevision)

	// generate a new key name
	uName, err := q.uniqueName()
	if err != nil {
		return &Item{}, err
	}
	uKey := path.Join(q.InFlight, uName)

	putReq := clientv3.OpPut(uKey, string(kv.Value))
	delReq := clientv3.OpDelete(key)
	response, err := q.kv.Txn(ctx).If(cmp).Then(putReq, delReq).Commit()
	if err != nil {
		return &Item{}, err
	}
	if response.Succeeded {
		return &Item{data: kv}, nil
	}
	return &Item{}, nil
}

// ensure that a waitPut also puts the item in the inflight lane and deletes it
// from the current work queue
func (q *Queue) waitPutEvent(ctx context.Context) (*clientv3.Event, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	wc := q.client.Watch(ctx, q.Work, clientv3.WithPrefix())
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
