package queue

import (
	"context"
	"errors"
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

type BackendIDGetter interface {
	GetBackendID() int64
}

// Item is a Queue item.
type Item struct {
	key   string
	value string
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
	return errors.New("not implemented")
}

// Nack returns the Item to the work queue and deletes it from the in-flight
// lane.
func (i *Item) Nack(ctx context.Context) error {
	return errors.New("not implemented")
}
