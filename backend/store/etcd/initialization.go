package etcd

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/sensu/sensu-go/backend/store"
)

const (
	initializationLockKey = ".initialized.lock"
	initializationKey     = ".initialized"
)

// StoreInitializer ...
type StoreInitializer struct {
	mutex  *concurrency.Mutex
	client *clientv3.Client

	ctx context.Context
}

// NewInitializer returns a new store initializer
func (store *etcdStore) NewInitializer() (store.Initializer, error) {
	client := store.client
	session, err := concurrency.NewSession(client) // TODO: move session into etcdStore?
	if err != nil {
		return nil, err
	}

	return &StoreInitializer{
		mutex:  concurrency.NewMutex(session, initializationLockKey),
		ctx:    context.TODO(),
		client: client,
	}, nil
}

// Lock mutex to avoid competing writes
func (s *StoreInitializer) Lock() error {
	if err := s.mutex.Lock(s.ctx); err != nil {
		return err
	}

	return nil
}

// IsInitialized checks the state of the .initialized key
func (s *StoreInitializer) IsInitialized() (bool, error) {
	r, err := s.client.Get(s.ctx, initializationKey)
	if err != nil {
		return false, err
	}

	return r.Count > 0, nil
}

// Finalize - set .initialized key
func (s *StoreInitializer) Finalize() error {
	if _, err := s.client.Put(s.ctx, initializationKey, "1"); err != nil {
		return err
	}

	return nil
}

// Unlock mutex
func (s *StoreInitializer) Unlock() error {
	return s.mutex.Unlock(s.ctx)
}
