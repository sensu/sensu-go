package queue

import (
	"bytes"
	"context"
	"encoding/binary"
	"path"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/backend/store/etcd"
)

// Define configuration here
const (
	queuePrefix = "queue"
)

// Queue is a FIFO queue that is backed by etcd.
// Queue is not durable. When an item is received by a client, it is deleted
// from etcd. Clients are responsible for handling the item, and there is no
// way to retrieve an item again once it has been Dequeued.
type Queue struct {
	client *clientv3.Client
	// combination of constant and value passed in
	// note: this should be resource namespaced as in keybuilder (use
	// keybuilder)
	Name string
	kv   clientv3.KV
}

// New returns an instance of Queue.
func New(name string, client *clientv3.Client) *Queue {
	queue := &Queue{
		client: client,
		Name:   path.Join(etcd.EtcdRoot, queuePrefix, name),
		kv:     clientv3.NewKV(client),
	}
	return queue
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

		now := time.Now().UnixNano()
		buf := new(bytes.Buffer)
		if err := binary.Write(buf, binary.BigEndian, now); err != nil {
			return err
		}
		key := path.Join(q.Name, buf.String())

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
	return nil
}

// Dequeue gets a value from the queue. It returns an error if the context
// is cancelled, the deadline exceeded, or if the client encounters an error.
func (q *Queue) Dequeue(ctx context.Context) (string, error) {
	prefix := q.Name
	response, err := q.client.Get(ctx, prefix, clientv3.WithFirstKey()...)
	if err != nil {
		return "", err
	}
	if len(response.Kvs) > 0 {
		kv, err := q.tryDelete(ctx, response.Kvs[0])
		if err != nil {
			return "", err
		}
		if kv != nil {
			return string(kv.Value), nil
		}
	}
	if response.More {
		// Need to retry, we are promised that there will be more.
		return q.Dequeue(ctx)
	}

	// Wait for the queue to receive an item
	event, err := q.waitPutEvent(ctx)
	if err != nil {
		return "", err
	}

	if event != nil {
		return string(event.Kv.Value), nil
	}

	return q.Dequeue(ctx)
}

func (q *Queue) tryDelete(ctx context.Context, kv *mvccpb.KeyValue) (*mvccpb.KeyValue, error) {
	key := string(kv.Key)
	cmp := clientv3.Compare(clientv3.ModRevision(key), "=", kv.ModRevision)
	req := clientv3.OpDelete(key)
	response, err := q.kv.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return nil, err
	}
	if response.Succeeded {
		return kv, nil
	}
	return nil, nil
}

func (q *Queue) waitPutEvent(ctx context.Context) (*clientv3.Event, error) {
	wc := q.client.Watch(ctx, q.Name, clientv3.WithPrefix())
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
