package js

import (
	"fmt"
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
// according to which assets are loaded into them. Javascript contexts which
// are not used for cacheMaxAge are disposed of.
type vmCache struct {
	vms  sync.Map
	done chan struct{}
}

type cacheValue struct {
	lastRead int64
	mu       sync.Mutex
	vm       *otto.Otto
}

func newVMCache() *vmCache {
	cache := &vmCache{
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
	c.vms.Range(func(key, value interface{}) bool {
		obj := value.(*cacheValue)
		obj.mu.Lock()
		defer obj.mu.Unlock()
		valueTime := time.Unix(obj.lastRead, 0)
		if time.Since(valueTime) > cacheMaxAge {
			c.vms.Delete(key)
		}
		return true
	})
}

// Acquire gets a VM from the cache. It is a copy of the cached value.
// The cache item is locked while in use.
// Users must call Dispose with the key after Acquire.
func (c *vmCache) Acquire(key string) *otto.Otto {
	val, ok := c.vms.Load(key)
	if !ok {
		return nil
	}
	obj := val.(*cacheValue)
	obj.mu.Lock()
	obj.lastRead = time.Now().Unix()
	return obj.vm.Copy()
}

// Dispose releases the lock on the cache item.
func (c *vmCache) Dispose(key string) {
	val, ok := c.vms.Load(key)
	if !ok {
		panic(fmt.Sprintf("dispose called on %q, but not found", key))
	}
	obj := val.(*cacheValue)
	obj.mu.Unlock()
}

// Init initializes the value in the cache.
func (c *vmCache) Init(key string, vm *otto.Otto) {
	val := &cacheValue{lastRead: time.Now().Unix(), vm: vm}
	c.vms.Store(key, val)
}
