package etcd

import (
	"context"
	"errors"

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

// List resources at a given prefix
func (s *Storage) List(ctx context.Context, prefix string, objsPtr interface{}) error {
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
