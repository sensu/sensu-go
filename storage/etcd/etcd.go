package etcd

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/sensu/sensu-go/internal/apis/meta"
	"github.com/sensu/sensu-go/runtime/codec"
	"github.com/sensu/sensu-go/storage"
	"go.etcd.io/etcd/clientv3"
)

// NewStorage returns new non-prefixed etcd store.
func NewStorage(client clientv3.KV, codec codec.Codec) storage.Store {
	return &Storage{
		client: client,
		codec:  codec,
	}
}

// Storage is a light wrapper around an etcd client.
type Storage struct {
	client clientv3.KV
	codec  codec.Codec
}

// Get a key from storage and deserialize it into objPtr.
func (s *Storage) Get(ctx context.Context, key string, objPtr interface{}) error {
	resp, err := s.client.Get(ctx, key)
	if err != nil {
		return err
	}

	if len(resp.Kvs) == 0 {
		return storage.ErrNotFound
	}

	if len(resp.Kvs) == 0 {
		return storage.ErrNotFound
	}

	v := resp.Kvs[0].Value

	return s.codec.Decode(v, objPtr)
}

// Create an object in the store.
func (s *Storage) Create(ctx context.Context, key string, objPtr interface{}) error {
	serialized, err := s.codec.Encode(objPtr)
	if err != nil {
		return err
	}

	txn := s.client.Txn(ctx).If(
		keyNotFound(key),
	).Then(
		put(key, string(serialized)),
	)

	resp, err := txn.Commit()
	if err != nil {
		return err
	}

	if !resp.Succeeded {
		return errors.New("could not create existing object")
	}

	return nil
}

// Update a key given with the serialized object.
func (s *Storage) Update(ctx context.Context, key string, objPtr interface{}) error {
	serialized, err := s.codec.Encode(objPtr)
	if err != nil {
		return err
	}

	txn := s.client.Txn(ctx).If(
		keyFound(key),
	).Then(
		put(key, string(serialized)),
	)

	resp, err := txn.Commit()
	if err != nil {
		return err
	}

	if !resp.Succeeded {
		return errors.New("could not update non-existent object")
	}

	return nil
}

// CreateOrUpdate creates an object in storage if it doesn't exist and otherwise
// updates an existing object.
func (s *Storage) CreateOrUpdate(ctx context.Context, key string, objPtr interface{}) error {
	serialized, err := s.codec.Encode(objPtr)
	if err != nil {
		return err
	}

	if _, err := s.client.Put(ctx, key, string(serialized)); err != nil {
		return err
	}

	return nil
}

// List all keys from storage under the provided prefix key and deserialize it
// into objsPtr.
func (s *Storage) List(key string, objsPtr interface{}) error {
	// Make sure the interface is a pointer, and that the element at this address
	// is a slice.
	// TODO: better validation and move that logic into its own package.
	// See https://github.com/kubernetes/apimachinery/blob/c6dd271be00615c6fa8c91fdf63381265a5f0e4e/pkg/conversion/helper.go#L27
	v := reflect.ValueOf(objsPtr)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer, but got %v type", v.Type())
	}
	if v.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("expected slice, but got %s", v.Elem().Kind())
	}
	v = v.Elem()

	opts := []clientv3.OpOption{
		clientv3.WithPrefix(),
	}
	resp, err := s.client.Get(context.TODO(), key, opts...)
	if err != nil {
		return err
	}

	for _, kv := range resp.Kvs {
		// Decode and append the value to v, which must be a slice.
		// See https://github.com/kubernetes/apiserver/blob/10d97565493b4eea44b1ef6c1b3fd47d2876a866/pkg/storage/etcd3/store.go#L786
		obj, ok := reflect.New(v.Type().Elem()).Interface().(meta.Object)
		if !ok {
			return fmt.Errorf("type assertion failed, got data of type %T, not meta.Object", v.Type().Elem())
		}
		if err := s.codec.Decode(kv.Value, obj); err != nil {
			return err
		}

		v.Set(reflect.Append(v, reflect.ValueOf(obj).Elem()))
	}

	return nil
}

func keyFound(key string) clientv3.Cmp {
	return clientv3.Compare(
		clientv3.CreateRevision(key),
		">",
		0,
	)
}

func keyNotFound(key string) clientv3.Cmp {
	return clientv3.Compare(
		clientv3.CreateRevision(key),
		"=",
		0,
	)
}

func put(key, value string) clientv3.Op {
	return clientv3.OpPut(key, value)
}
