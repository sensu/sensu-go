package v2

import (
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
func (p *Proxy) CreateOrUpdate(req ResourceRequest, wrapper Wrapper) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.CreateOrUpdate(req, wrapper)
}

// UpdateIfExists updates the resource with the wrapped resource, but only
// if it already exists in the store.
func (p *Proxy) UpdateIfExists(req ResourceRequest, wrapper Wrapper) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.UpdateIfExists(req, wrapper)
}

// CreateIfNotExists writes the wrapped resource to the store, but only if
// it does not already exist.
func (p *Proxy) CreateIfNotExists(req ResourceRequest, wrapper Wrapper) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.CreateIfNotExists(req, wrapper)
}

// Get gets a wrapped resource from the store.
func (p *Proxy) Get(req ResourceRequest) (Wrapper, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.Get(req)
}

// Delete deletes a resource from the store.
func (p *Proxy) Delete(req ResourceRequest) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.Delete(req)
}

// List lists all resources specified by the resource request, and the
// selection predicate.
func (p *Proxy) List(req ResourceRequest, pred *store.SelectionPredicate) (WrapList, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.List(req, pred)
}

// Exists returns true if the resource indicated by the request exists
func (p *Proxy) Exists(req ResourceRequest) (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.Exists(req)
}

// Patch patches the resource given in the request
func (p *Proxy) Patch(req ResourceRequest, wrapper Wrapper, patcher patch.Patcher, cond *store.ETagCondition) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.impl.Patch(req, wrapper, patcher, cond)
}
