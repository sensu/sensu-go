package etcd

import (
	"context"
	"fmt"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/util/retry"
)

var (
	backendIDKeyPrefix         = store.NewKeyBuilder("backends").Build()
	backendIDLeasePeriod int64 = 60
	minRetryLeaseDelay         = time.Second
	maxRetryLeaseDelay         = time.Minute
	retryLeaseTimeout          = time.Hour
	retryLeaseMultiplier       = 2.0
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
	}
	// Wait until the backend ID has been created
	getter.wg.Add(1)

	// Start the async worker that populates the backend ID
	go getter.retryAcquireLease()

	// Wait until the worker has acquired a backend ID
	getter.wg.Wait()

	return getter
}

func leaseRetryBackoff() *retry.ExponentialBackoff {
	return &retry.ExponentialBackoff{
		InitialDelayInterval: minRetryLeaseDelay,
		MaxDelayInterval:     maxRetryLeaseDelay,
		MaxElapsedTime:       retryLeaseTimeout,
		Multiplier:           retryLeaseMultiplier,
	}
}

func (b *BackendIDGetter) retryAcquireLease() {
	backoff := leaseRetryBackoff()
	for {
		var ch <-chan *clientv3.LeaseKeepAliveResponse
		err := backoff.Retry(func(retries int) (bool, error) {
			var err error
			var id int64
			id, ch, err = b.getLease()
			if err != nil {
				logger.WithError(err).Error("error generating backend ID")
				if err := b.ctx.Err(); err != nil {
					return true, err
				}
				return false, nil
			}
			atomic.StoreInt64(&b.id, id)
			b.wg.Done()
			return true, nil
		})
		if err != nil && err != b.ctx.Err() {
			// Crash at this point. The system could not acquire a lease for
			// retryLeaseTimeout duration.
			panic(fmt.Sprintf("couldn't acquire an etcd lease for %v", retryLeaseTimeout))
		}
		for resp := range ch {
			if resp.ID == clientv3.NoLease {
				break
			}
		}
		b.wg.Add(1)
	}
}

func (b *BackendIDGetter) getLease() (int64, <-chan *clientv3.LeaseKeepAliveResponse, error) {
	// Grant a lease for 60 seconds
	resp, err := b.client.Grant(b.ctx, backendIDLeasePeriod)
	if err != nil {
		return 0, nil, fmt.Errorf("error creating backend ID: error granting lease: %s", err)
	}

	// Register the backend's lease - this is for clients that need to be
	// able to send specific backends messages
	value := fmt.Sprintf("%x", resp.ID)
	key := path.Join(backendIDKeyPrefix, value)
	_, err = b.client.Put(b.ctx, key, value, clientv3.WithLease(resp.ID))
	if err != nil {
		return 0, nil, fmt.Errorf("error creating backend ID: error creating key: %s", err)
	}

	// Keep the lease alive
	ch, err := b.client.KeepAlive(b.ctx, resp.ID)

	return int64(resp.ID), ch, err
}
