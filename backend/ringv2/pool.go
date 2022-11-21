package ringv2

import (
	"sync"
)

type NewFunc func(path string) Interface

// RingPool is a pool for rings. It uses a NewFunc to create new rings when
// needed. RingPool supercedes Pool, by using the Interface type instead of a
// *Ring.
type RingPool struct {
	newf  NewFunc
	rings map[string]Interface
	mu    sync.Mutex
}

// NewRingPool creates a new RingPool.
func NewRingPool(fn NewFunc) *RingPool {
	return &RingPool{
		newf:  fn,
		rings: make(map[string]Interface),
	}
}

// Get gets the ring corresponding to the given path.
func (r *RingPool) Get(path string) Interface {
	r.mu.Lock()
	defer r.mu.Unlock()
	ring, ok := r.rings[path]
	if ok {
		return ring
	}
	ring = r.newf(path)
	r.rings[path] = ring
	return ring
}

// Del deletes the ring corresponding to the given path.
func (r *RingPool) Del(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rings, path)
}

// SetNewFunc sets the newer function for the ring pool. It results in the
// pool being cleared.
func (r *RingPool) SetNewFunc(fn NewFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.newf = fn
	r.rings = make(map[string]Interface, len(r.rings))
}
