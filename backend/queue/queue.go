package queue

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/sensu/sensu-go/backend/store"
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

// Client interface to queue backend
type Client interface {
	// Enqueue Item
	Enqueue(context.Context, Item) error
	// Reserve an Item from the queue
	// Blocks until an item can be reserved or the context expires
	Reserve(ctx context.Context, queue string) (Reservation, error)
}

// Item is a Queue Item.
type Item struct {
	// ID of queue item. Ignored on Enqueue
	ID string
	// Queue name
	Queue string
	// Value of queue item
	Value []byte
}

// Reservation for a Queue Item.
// Must be either Ack'd or Nack'd.
type Reservation interface {
	Item() Item
	// Ack the Item was handeled and can be deleted.
	Ack(context.Context) error
	// Nack the Item could not be handled. Return to the queue.
	Nack(context.Context) error
}

// ClusteredQueue is a Client wrapper meant to function as a simple pub/sub
// across all active sensu backends in a cluster.
// Given a queueName, ClusteredQueue Enqueues to queueName/{{backendID}} for
// each backend, and Reserves only from queueName/{{this backendID}}
type ClusteredQueue struct {
	// Client underlying queue implementation
	Client Client
	// GetBackendID gets the current backend process's id.
	// TODO(ck): just hostname per backend/opc.go?
	GetBackendID func() string
	// GetAllBackendIDs returns the set of currently present backends.
	// TODO(ck): Find source for this. The OPC table makes sense as a source of
	// truth for this, but we need more flexible query mechanism.
	GetAllBackendIDs func(context.Context) ([]string, error)
}

// Enqueue Item to "{{item.Queue}}/{{backend id}}" for each backend id
func (q *ClusteredQueue) Enqueue(ctx context.Context, item Item) error {
	backendIDs, err := q.GetAllBackendIDs(ctx)
	if err != nil {
		return fmt.Errorf("error getting backend IDs: %w", err)
	}
	baseQueue := item.Queue
	for _, id := range backendIDs {
		item.Queue = path.Join(baseQueue, id)
		if err := q.Client.Enqueue(ctx, item); err != nil {
			return fmt.Errorf("error enqueuing item to queue %s: %w", item.Queue, err)
		}
	}
	return nil
}

// Reserve Item from "{{queue}}/{{backend id}}"
func (q *ClusteredQueue) Reserve(ctx context.Context, queue string) (Reservation, error) {
	return q.Client.Reserve(ctx, path.Join(queue, q.GetBackendID()))
}
