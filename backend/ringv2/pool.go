package ringv2

import (
	"context"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/robfig/cron"
)

// Pool is a pool of rings. It exists to help users avoid creating too many
// watchers.
type Pool struct {
	client *clientv3.Client
	rings  map[string]*ringProxy
	mu     sync.Mutex
}

// NewPool creates a new Pool.
func NewPool(client *clientv3.Client) *Pool {
	return &Pool{
		client: client,
		rings:  make(map[string]*ringProxy),
	}
}

type ringProxy struct {
	ring *Ring
	ch   *<-chan Event
	sync.Mutex
}

func (r *ringProxy) Add(ctx context.Context, value string, keepalive int64) error {
	return r.ring.Add(ctx, value, keepalive)
}

func (r *ringProxy) Remove(ctx context.Context, value string) error {
	return r.ring.Remove(ctx, value)
}

func (r *ringProxy) SetInterval(ctx context.Context, seconds int64) error {
	return r.ring.SetInterval(ctx, seconds)
}

func (r *ringProxy) SetCron(schedule cron.Schedule) {
	r.ring.SetCron(schedule)
}

// Watch returns a watcher for the ring. The first result of an underlying
// call to Watch is cached, and returned on subsequent invocations. This means
// that subsequent calls to Watch will not use the provided context.
func (r *ringProxy) Watch(ctx context.Context, n int) <-chan Event {
	r.Lock()
	defer r.Unlock()
	if r.ch == nil {
		wc := r.ring.Watch(ctx, n)
		r.ch = &wc
		go func() {
			<-ctx.Done()
			r.Lock()
			defer r.Unlock()
			r.ch = nil
		}()
	}
	return *r.ch
}

// Get gets a ring from the pool.
func (p *Pool) Get(path string) Interface {
	p.mu.Lock()
	ring, ok := p.rings[path]
	p.mu.Unlock()
	if ok {
		return ring
	}
	ring = &ringProxy{
		ring: New(p.client, path),
	}
	p.mu.Lock()
	p.rings[path] = ring
	p.mu.Unlock()
	return ring
}
