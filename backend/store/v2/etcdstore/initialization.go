package etcdstore

import (
	"context"
	"fmt"
	"path"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/store"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

const (
	initializationLockKey = ".initialized.lock"
	initializationKey     = ".initialized"
)

// StoreInitializer ...
type StoreInitializer struct {
	mutex *concurrency.Mutex

	session *concurrency.Session
	client  *clientv3.Client
}

// NewInitializer returns a new store initializer
func (s *Store) NewInitializer(ctx context.Context) (store.Initializer, error) {
	client := s.client

	// Create a lease to associate with the lock
	resp, err := client.Grant(ctx, 2)
	etcd.LeaseOperationsCounter.WithLabelValues("init", etcd.LeaseOperationTypeGrant, etcd.LeaseStatusFor(err)).Inc()
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd lease: %w", err)
	}

	session, err := concurrency.NewSession(client, concurrency.WithLease(resp.ID)) // TODO: move session into etcdStore?
	etcd.LeaseOperationsCounter.WithLabelValues("init", etcd.LeaseOperationTypePut, etcd.LeaseStatusFor(err)).Inc()
	if err != nil {
		return nil, fmt.Errorf("failed to start etcd session: %w", err)
	}

	return &StoreInitializer{
		mutex:   concurrency.NewMutex(session, initializationLockKey),
		session: session,
		client:  client,
	}, nil
}

// Lock mutex to avoid competing writes
func (s *StoreInitializer) Lock(ctx context.Context) error {
	return s.mutex.Lock(ctx)
}

// IsInitialized checks the state of the .initialized key
func (s *StoreInitializer) IsInitialized(ctx context.Context) (bool, error) {
	r, err := s.client.Get(ctx, initializationPath())
	if err != nil {
		return false, err
	}

	// if there is no result, test the fallback and return using same logic
	if len(r.Kvs) == 0 {
		fallback, err := s.client.Get(ctx, initializationKey)
		if err != nil {
			return false, err
		} else {
			return fallback.Count > 0, nil
		}
	}

	return r.Count > 0, nil
}

// FlagAsInitialized - set .initialized key
func (s *StoreInitializer) FlagAsInitialized(ctx context.Context) error {
	_, err := s.client.Put(ctx, initializationPath(), "1")
	return err
}

// Close session & unlock
func (s *StoreInitializer) Close(ctx context.Context) error {
	if err := s.mutex.Unlock(ctx); err != nil {
		return err
	}

	if err := s.session.Close(); err != nil {
		return err
	}
	<-s.session.Done()

	return nil
}

func initializationPath() string {
	return path.Join("/sensu.io", initializationKey)
}
