package etcd

import (
	"context"
	"errors"
	"fmt"
	"path"
	"sync"
	"sync/atomic"

	"github.com/sensu/sensu-go/backend/store"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	backendIDKeyPrefix         = store.NewKeyBuilder("backends").Build()
	backendIDLeasePeriod int64 = 60
)

// BackendIDGetterClient represents the dependencies for BackendIDGetter.
type BackendIDGetterClient interface {
	Grant(ctx context.Context, ttl int64) (*clientv3.LeaseGrantResponse, error)
	KeepAlive(ctx context.Context, id clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error)
	Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)
}

// BackendIDGetter is a type that facilitates identifying a sensu backend.
type BackendIDGetter struct {
	id     int64
	wg     sync.WaitGroup
	ctx    context.Context
	client BackendIDGetterClient
	errors chan error
}

func (b *BackendIDGetter) GetBackendID() int64 {
	// The backend ID might have been invalidated by a lease that was allowed
	// to expire. If that's the case, wait for a new lease to be granted.
	b.wg.Wait()

	return atomic.LoadInt64(&b.id)
}

// NewBackendIDGetter creates a new BackendIDGetter. It uses a context that
// should be valid for the life of the application, to pass to etcd.
// It requires a BackendIDGetterClient, which users can provide by using
// an etcd *clientv3.Client.
func NewBackendIDGetter(ctx context.Context, client BackendIDGetterClient) *BackendIDGetter {
	getter := &BackendIDGetter{
		client: client,
		ctx:    ctx,
		errors: make(chan error, 1),
	}
	// Wait until the backend ID has been created
	getter.wg.Add(1)

	// Start the async worker that populates the backend ID
	go getter.keepAliveLease(ctx)

	// Wait until the worker has acquired a backend ID
	getter.wg.Wait()

	return getter
}

func (b *BackendIDGetter) keepAliveLease(ctx context.Context) {
	id, ch, err := b.getLease()
	if err != nil {
		if err != ctx.Err() {
			logger.WithError(err).Error("error generating backend ID")
			b.errors <- err
		}
		return
	}
	atomic.StoreInt64(&b.id, id)
	b.wg.Done()
	for {
		select {
		case resp, ok := <-ch:
			if !ok {
				if ctx.Err() == nil {
					b.errors <- errors.New("keeplive channel closed")
				}
				return
			}
			LeaseOperationsCounter.WithLabelValues(LeaseOperationTypeKeepalive, LeaseOperationStatusAlive).Inc()
			if resp.ID == clientv3.NoLease {
				// I believe this to be impossible
				b.errors <- errors.New("no lease")
			}
		case <-ctx.Done():
			return
		}
	}
}

func (b *BackendIDGetter) getLease() (int64, <-chan *clientv3.LeaseKeepAliveResponse, error) {
	// Grant a lease for 60 seconds
	resp, err := b.client.Grant(b.ctx, backendIDLeasePeriod)
	LeaseOperationsCounter.WithLabelValues(LeaseOperationTypeGrant, LeaseStatusFor(err)).Inc()
	if err != nil {
		return 0, nil, fmt.Errorf("error creating backend ID: error granting lease: %s", err)
	}
	leaseID := resp.ID

	// Register the backend's lease - this is for clients that need to be
	// able to send specific backends messages
	value := fmt.Sprintf("%x", leaseID)
	key := path.Join(backendIDKeyPrefix, value)
	_, err = b.client.Put(b.ctx, key, value, clientv3.WithLease(leaseID))
	LeaseOperationsCounter.WithLabelValues(LeaseOperationTypePut, LeaseStatusFor(err)).Inc()
	if err != nil {
		return 0, nil, fmt.Errorf("error creating backend ID: error creating key: %s", err)
	}

	// Keep the lease alive
	ch, err := b.client.KeepAlive(b.ctx, leaseID)
	LeaseOperationsCounter.WithLabelValues(LeaseOperationTypeKeepalive, LeaseStatusFor(err)).Inc()

	return int64(leaseID), ch, err
}

func (b *BackendIDGetter) Stop() error {
	// no-op as we're controlled by the context
	return nil
}

func (b *BackendIDGetter) Start() error {
	// no-op as we start on New
	return nil
}

func (b *BackendIDGetter) Err() <-chan error {
	return b.errors
}

func (b *BackendIDGetter) Name() string {
	return "backend_id_getter"
}
