package js

import (
	"sync"

	time "github.com/echlebek/timeproxy"
	"github.com/robertkrimen/otto"
)

const (
	// cacheMaxAge is the maximum amount of time to leave an unused
	// item in the cache for.
	cacheMaxAge = time.Hour

	// cacheReapInterval is the amount to sleep in the cache reaper
	cacheReapInterval = time.Minute
)

// vmCache provides an internal mechanism for caching javascript contexts
// according to which assets are loaded into them. Javascrip contexts which
// are not used for cacheMaxAge are disposed of.
type vmCache struct {
	vms  map[string]*cacheValue
	done chan struct{}
	sync.Mutex
}

type cacheValue struct {
	lastRead int64
	vm       *otto.Otto
}

func newVMCache() *vmCache {
	cache := &vmCache{
		vms:  make(map[string]*cacheValue),
		done: make(chan struct{}),
	}
	go cache.reapLoop()
	return cache
}

func (c *vmCache) Close() {
	close(c.done)
}

func (c *vmCache) reapLoop() {
	// reap old cache items
	ticker := time.NewTicker(cacheReapInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.reap()
		}
	}
}

func (c *vmCache) reap() {
	c.Lock()
	for k, v := range c.vms {
		valueTime := time.Unix(v.lastRead, 0)
		if valueTime.Before(time.Now().Add(-cacheMaxAge)) {
			delete(c.vms, k)
		}
	}
	defer c.Unlock()
}

// Acquire gets a VM from the cache. It is a copy of the cached value.
func (c *vmCache) Acquire(key string) *otto.Otto {
	c.Lock()
	defer c.Unlock()
	val, ok := c.vms[key]
	if !ok {
		return nil
	}
	if val.vm == nil {
		return nil
	}
	return val.vm.Copy()
}

// Init initializes the value in the cache.
func (c *vmCache) Init(key string, vm *otto.Otto) {
	c.Lock()
	defer c.Unlock()
	val := &cacheValue{lastRead: time.Now().Unix(), vm: vm}
	c.vms[key] = val
}
