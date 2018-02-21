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
	mutex *concurrency.Mutex
	ctx   context.Context

	session *concurrency.Session
	client  *clientv3.Client
}

// NewInitializer returns a new store initializer
func (store *Store) NewInitializer() (store.Initializer, error) {
	client := store.client
	session, err := concurrency.NewSession(client) // TODO: move session into etcdStore?
	if err != nil {
		return nil, err
	}

	return &StoreInitializer{
		mutex:   concurrency.NewMutex(session, initializationLockKey),
		session: session,
		client:  client,
		ctx:     context.TODO(),
	}, nil
}

// Lock mutex to avoid competing writes
func (s *StoreInitializer) Lock() error {
	return s.mutex.Lock(s.ctx)
}

// IsInitialized checks the state of the .initialized key
func (s *StoreInitializer) IsInitialized() (bool, error) {
	r, err := s.client.Get(s.ctx, initializationKey)
	if err != nil {
		return false, err
	}

	return r.Count > 0, nil
}

// FlagAsInitialized - set .initialized key
func (s *StoreInitializer) FlagAsInitialized() error {
	_, err := s.client.Put(s.ctx, initializationKey, "1")
	return err
}

// Close session & unlock
func (s *StoreInitializer) Close() error {
	if err := s.mutex.Unlock(s.ctx); err != nil {
		return err
	}

	if err := s.session.Close(); err != nil {
		return err
	}
	<-s.session.Done()

	return nil
}
