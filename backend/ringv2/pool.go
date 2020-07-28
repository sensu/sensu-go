package ringv2

import (
	"sync"

	"go.etcd.io/etcd/clientv3"
)

// Pool is a pool of rings. It exists to help users avoid creating too many
// watchers.
type Pool struct {
	client *clientv3.Client
	rings  map[string]*Ring
	mu     sync.Mutex
}

// NewPool creates a new Pool.
func NewPool(client *clientv3.Client) *Pool {
	return &Pool{
		client: client,
		rings:  make(map[string]*Ring),
	}
}

// Get gets a ring from the pool.
func (p *Pool) Get(path string) *Ring {
	p.mu.Lock()
	ring, ok := p.rings[path]
	p.mu.Unlock()
	if ok {
		return ring
	}
	ring = New(p.client, path)
	p.mu.Lock()
	p.rings[path] = ring
	p.mu.Unlock()
	return ring
}
