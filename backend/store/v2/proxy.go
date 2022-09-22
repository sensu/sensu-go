package v2

import (
	"context"
	"sync"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
)

type Proxy struct {
	mu   sync.RWMutex
	impl Interface
}

func (p *Proxy) UpdateStore(store Interface) {
	if store == p {
		panic("UpdateStore called with itself as argument")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.impl = store
}

// CreateOrUpdate creates or updates the wrapped resource.
func (p *Proxy) CreateOrUpdate(ctx context.Context, req ResourceRequest, wrapper Wrapper) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.CreateOrUpdate(ctx, req, wrapper)
}

// UpdateIfExists updates the resource with the wrapped resource, but only
// if it already exists in the store.
func (p *Proxy) UpdateIfExists(ctx context.Context, req ResourceRequest, wrapper Wrapper) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.UpdateIfExists(ctx, req, wrapper)
}

// CreateIfNotExists writes the wrapped resource to the store, but only if
// it does not already exist.
func (p *Proxy) CreateIfNotExists(ctx context.Context, req ResourceRequest, wrapper Wrapper) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.CreateIfNotExists(ctx, req, wrapper)
}

// Get gets a wrapped resource from the store.
func (p *Proxy) Get(ctx context.Context, req ResourceRequest) (Wrapper, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.Get(ctx, req)
}

// Delete deletes a resource from the store.
func (p *Proxy) Delete(ctx context.Context, req ResourceRequest) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.Delete(ctx, req)
}

// List lists all resources specified by the resource request, and the
// selection predicate.
func (p *Proxy) List(ctx context.Context, req ResourceRequest, pred *store.SelectionPredicate) (WrapList, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.List(ctx, req, pred)
}

// Exists returns true if the resource indicated by the request exists
func (p *Proxy) Exists(ctx context.Context, req ResourceRequest) (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.Exists(ctx, req)
}

// Patch patches the resource given in the request
func (p *Proxy) Patch(ctx context.Context, req ResourceRequest, wrapper Wrapper, patcher patch.Patcher, cond *store.ETagCondition) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.Patch(ctx, req, wrapper, patcher, cond)
}

// Watch sets up a watcher that responds to updates to the given key or
// keyspace indicated by the ResourceRequest.
func (p *Proxy) Watch(ctx context.Context, req ResourceRequest) <-chan []WatchEvent {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.Watch(ctx, req)
}

// Initialize sets up a cluster with the default resources & config.
func (p *Proxy) Initialize(ctx context.Context, fn InitializeFunc) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.Initialize(ctx, fn)
}
