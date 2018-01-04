package queue

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store/etcd"
)

// Notes from chat about queues:
// Queues such as Rabbit and Beanstalkd have message safety/durability
// implemented as "ready" and "reserved" state. We may not need to add implement
// that here due to the following:
// 	- queue use will be limited to the backend
// 	- the backend is responsible for sending check requests to the agent
//  - if the backend fails (due to a partition or other failure) only its
//  currently processed item would be lost
//  - if the check fails to execute due to an agent failure, a keepalive event,
//  check ttl failure, or both will be created.
//  - the backend is only responsible for the message delivery to the agent, not
//  that the agent handled the message.

// Need to define functions for:
// - add item to queue
// - remove item from queue
// - delete item from queue (either part of remove or a separate action if we
// want durability)

// Define configuration here
const (
	queuePrefix = "queue"
)

// Queue defines an etcd queue
type Queue struct {
	Client *clientv3.Client
	// combination of constant and value passed in
	// note: this should be resource namespaced as in keybuilder (use
	// keybuilder)
	Name string
}

// New returns an instance of Queue.
func New(name string, client *clientv3.Client) *Queue {
	queue := &Queue{
		Client: client,
		Name:   path.Join(etcd.EtcdRoot, queuePrefix, name),
	}
	return queue
}

// Enqueue adds a new value to the queue. It returns an error if the context is
// canceled, the deadline exceeded, or if the client encounters an error.
func (q *Queue) Enqueue(ctx context.Context, value string) error {
	return q.tryPut(ctx, value)
}

func (q *Queue) tryPut(ctx context.Context, value string) error {
	kv := clientv3.NewKV(q.Client)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		key := path.Join(q.Name, fmt.Sprintf("%d", time.Now().UnixNano()))

		cmp := clientv3.Compare(clientv3.Version(key), "=", 0)
		req := clientv3.OpPut(key, value)
		response, err := kv.Txn(ctx).If(cmp).Then(req).Commit()
		if err == nil && response.Succeeded {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// dequeue

// delete?
