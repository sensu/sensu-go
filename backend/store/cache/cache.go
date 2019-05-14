package cache

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/coreos/etcd/clientv3"
	time "github.com/echlebek/timeproxy"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd"
)

// cacheKey is used to uniquely identify cache entries
type cacheKey struct {
	Name      string
	Namespace string
}

func getCacheKey(resource corev2.Resource) cacheKey {
	return cacheKey{
		Namespace: resource.GetObjectMeta().Namespace,
		Name:      resource.GetObjectMeta().Name,
	}
}

type cacheSlice []corev2.Resource

func (s cacheSlice) Len() int {
	return len(s)
}

func (s cacheSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func resourceLT(x, y corev2.Resource) bool {
	if x == nil {
		return true
	}
	if y == nil {
		return false
	}
	if x.GetObjectMeta().Namespace == y.GetObjectMeta().Namespace {
		return x.GetObjectMeta().Name < y.GetObjectMeta().Name
	}
	return x.GetObjectMeta().Namespace < y.GetObjectMeta().Namespace
}

func (s cacheSlice) Less(i, j int) bool {
	return resourceLT(s[i], s[j])
}

type cacheWatcher struct {
	ctx context.Context
	ch  chan struct{}
}

// ResourceCacher is a cache of resources. The cache uses a watcher on a certain
// type of resources in order to keep itself up to date. Cache resources can be
// efficiently retrieved from the cache by namespace.
type ResourceCacher struct {
	watcher      <-chan store.WatchEventResource
	mapCache     map[cacheKey]corev2.Resource
	sliceCache   []corev2.Resource
	updates      []store.WatchEventResource
	cacheMu      sync.Mutex
	watchers     []cacheWatcher
	watchersMu   sync.Mutex
	resourceType reflect.Type
}

// New creates a new Resource cache. It retrieves all resources from the
// store on creation.
func New(ctx context.Context, client *clientv3.Client, keyBuilder etcd.KeyBuilderFn, elem interface{}) (*ResourceCacher, error) {
	elemType := reflect.TypeOf(elem)
	sliceOfElem := reflect.SliceOf(elemType)
	ptr := reflect.New(sliceOfElem)
	ptr.Elem().Set(reflect.MakeSlice(sliceOfElem, 0, 10))

	err := etcd.List(ctx, client, keyBuilder, ptr.Interface(), &store.SelectionPredicate{})
	if err != nil {
		return nil, fmt.Errorf("error creating ResourceCacher: %s", err)
	}

	elemSlice := reflect.Indirect(ptr)
	resources := make([]corev2.Resource, elemSlice.Len())
	for i := 0; i < elemSlice.Len(); i++ {
		resource := elemSlice.Index(i).Interface().(corev2.Resource)
		resources[i] = resource
	}

	mapCache := make(map[cacheKey]corev2.Resource, len(resources))
	for _, resource := range resources {
		mapCache[getCacheKey(resource)] = resource
	}
	cache := &ResourceCacher{
		sliceCache: resources,
		mapCache:   mapCache,
		watcher:    etcd.GetResourceWatcher(ctx, client, keyBuilder(ctx, ""), elemType),
	}
	go cache.start(ctx)
	return cache, nil
}

// Watch allows cache users to get notified when the cache has new values.
// When the context is canceled, the channel will be closed.
func (c *ResourceCacher) Watch(ctx context.Context) <-chan struct{} {
	watcher := cacheWatcher{
		ctx: ctx,
		ch:  make(chan struct{}, 1),
	}
	c.watchersMu.Lock()
	c.watchers = append(c.watchers, watcher)
	c.watchersMu.Unlock()
	return watcher.ch
}

func (c *ResourceCacher) notifyWatchers() {
	c.watchersMu.Lock()
	defer c.watchersMu.Unlock()
	deletes := map[int]struct{}{}
	for i, watcher := range c.watchers {
		if err := watcher.ctx.Err(); err != nil {
			deletes[i] = struct{}{}
			continue
		}
		select {
		case watcher.ch <- struct{}{}:
		default:
			// if there is already a notification in the buffer, don't send
			// another.
		}
	}
	newWatchers := make([]cacheWatcher, 0, len(c.watchers))
	for i, w := range c.watchers {
		if _, ok := deletes[i]; !ok {
			newWatchers = append(newWatchers, w)
		}
	}
	c.watchers = newWatchers
}

func (c *ResourceCacher) start(ctx context.Context) {
	// 1s is the minimum scheduling interval, and so is the rate that
	// the cache will update at.
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-c.watcher:
			c.updates = append(c.updates, event)
		case <-ticker.C:
			if len(c.updates) > 0 {
				c.updateCache()
				c.notifyWatchers()
			}
		}
	}
}

// Get gets all resources in a namespace.
func (c *ResourceCacher) Get(namespace string) []corev2.Resource {
	c.cacheMu.Lock()
	cache := c.sliceCache
	c.cacheMu.Unlock()
	start := sort.Search(len(cache), func(i int) bool {
		return cache[i].GetObjectMeta().Namespace >= namespace
	})
	endNS := namespace + string(rune(0))
	stop := sort.Search(len(cache), func(i int) bool {
		return cache[i].GetObjectMeta().Namespace >= endNS
	})
	if stop > len(cache) {
		stop = len(cache)
	}
	return cache[start:stop]
}

func (c *ResourceCacher) updateCache() {
	for _, event := range c.updates {
		resource := event.Resource
		if resource == nil {
			logger.Error("nil resource in watch event")
			continue
		}
		key := getCacheKey(resource)
		switch event.Action {
		case store.WatchCreate, store.WatchUpdate:
			c.mapCache[key] = resource
		case store.WatchDelete:
			delete(c.mapCache, key)
		default:
			logger.Error("error in resource watcher")
		}
	}
	c.updates = nil
	newSliceCache := make([]corev2.Resource, 0, len(c.mapCache))
	for _, v := range c.mapCache {
		newSliceCache = append(newSliceCache, v)
	}
	sort.Sort(cacheSlice(newSliceCache))
	c.cacheMu.Lock()
	c.sliceCache = newSliceCache
	c.cacheMu.Unlock()
}
