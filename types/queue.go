package types

import "context"

// Queue is the interface of a queue. Queue's methods are atomic
// and goroutine-safe.
type Queue interface {
	// Enqueue adds a new item to the queue. It returns any error that
	// was encountered in doing so, or if the context is cancelled.
	Enqueue(ctx context.Context, value string) error

	// Dequeue gets an Item from the queue. It returns the Item and any
	// error encountered, or if the context is cancelled.
	Dequeue(ctx context.Context) (QueueItem, error)
}

// QueueItem represents an item retrieved from a Queue.
type QueueItem interface {
	// Value is the item's underlying value.
	Value() string

	// Ack acks the item. The item will no longer be stored.
	Ack(context.Context) error

	// Nack nacks the item. The item will return to the Queue.
	Nack(context.Context) error
}

// QueueGetter is a type that provides a way to get a Queue.
type QueueGetter interface {
	GetQueue(path ...string) Queue
}
