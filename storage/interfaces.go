package storage

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a key is not found in the store.
var ErrNotFound = errors.New("key not found")

// A Store provides the methods necessary for interacting with
// objects in some form of storage.
type Store interface {
	Create(ctx context.Context, key string, obj interface{}) error
	Update(ctx context.Context, key string, obj interface{}) error
	CreateOrUpdate(ctx context.Context, key string, obj interface{}) error
	Get(ctx context.Context, key string, obj interface{}) error
	List(ctx context.Context, prefix string, objs interface{}) error
}
