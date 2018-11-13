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

	// cacheMaxConcurrency is the maximum number of concurrent accesses
	// available for a given key.
	cacheMaxConcurrency = 8
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
	vms      chan *otto.Otto
}

func newCacheValue() *cacheValue {
	return &cacheValue{
		vms: make(chan *otto.Otto, cacheMaxConcurrency),
	}
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

// Acquire gets a VM from the cache. When the user is finished with it, it
// should be returned to the cache with Release.
func (c *vmCache) Acquire(key string) *otto.Otto {
	c.Lock()
	defer c.Unlock()
	val, ok := c.vms[key]
	if !ok {
		return nil
	}
	return <-val.vms
}

// Init initializes the value in the cache, creating cacheMaxConcurrency copies
// of it.
func (c *vmCache) Init(key string, vm *otto.Otto) {
	c.Lock()
	defer c.Unlock()
	val := newCacheValue()
	c.vms[key] = val
	// Fill the cache with copies of the VM
	for i := 0; i < cacheMaxConcurrency; i++ {
		val.vms <- vm.Copy()
	}
	val.lastRead = time.Now().Unix()
}

// Release returns a VM to the cache.
func (c *vmCache) Release(key string, vm *otto.Otto) {
	c.Lock()
	defer c.Unlock()
	c.vms[key].vms <- vm
}
