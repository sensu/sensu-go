package cache

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/types/dynamic"
)

// key is used to uniquely identify cache entries
type key struct {
	name      string
	namespace string
}

// Value contains a cached value, and its synthesized companion.
type Value struct {
	Resource corev2.Resource
	Synth    interface{}
}

func getCacheKey(resource corev2.Resource) key {
	return key{
		namespace: resource.GetObjectMeta().Namespace,
		name:      resource.GetObjectMeta().Name,
	}
}

func getCacheValue(resource corev2.Resource, synthesize bool) Value {
	v := Value{Resource: resource}
	if synthesize {
		v.Synth = dynamic.Synthesize(resource)
	}
	return v
}

// MakeSliceCache ...
func MakeSliceCache(resources []corev2.Resource, synthesize bool) []Value {
	cache := make([]Value, len(resources))
	for i := range cache {
		cache[i] = getCacheValue(resources[i], synthesize)
	}
	return cache
}

type resourceSlice []Value

func (s resourceSlice) Len() int {
	return len(s)
}

func (s resourceSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s resourceSlice) Less(i, j int) bool {
	return resourceLT(s[i], s[j])
}

func resourceLT(x, y Value) bool {
	if x.Resource == nil {
		return true
	}
	if y.Resource == nil {
		return false
	}
	if x.Resource.GetObjectMeta().Namespace == y.Resource.GetObjectMeta().Namespace {
		return x.Resource.GetObjectMeta().Name < y.Resource.GetObjectMeta().Name
	}
	return x.Resource.GetObjectMeta().Namespace < y.Resource.GetObjectMeta().Namespace
}

type cacheWatcher struct {
	ctx context.Context
	ch  chan struct{}
}

// Resource is a cache of resources. The cache uses a watcher on a certain
// type of resources in order to keep itself up to date. Cache resources can be
// efficiently retrieved from the cache by namespace.
type Resource struct {
	watcher    <-chan store.WatchEventResource
	mapCache   map[key]Value
	sliceCache []Value
	updates    []store.WatchEventResource
	cacheMu    sync.Mutex
	watchers   []cacheWatcher
	watchersMu sync.Mutex
	synthesize bool
}

// New creates a new resource cache. It retrieves all resources from the
// store on creation.
func New(ctx context.Context, client *clientv3.Client, resource corev2.Resource, synthesize bool) (*Resource, error) {
	// Get the type of the resource and create a slice type of []type
	typeOfResource := reflect.TypeOf(resource)
	sliceOfResource := reflect.SliceOf(typeOfResource)
	// Create a pointer to our slice type and then set the slice value
	ptr := reflect.New(sliceOfResource)
	ptr.Elem().Set(reflect.MakeSlice(sliceOfResource, 0, 0))

	// Get a keybuilderFunc for this resource
	keyBuilderFunc := func(ctx context.Context, name string) string {
		return store.NewKeyBuilder(resource.StorePrefix()).WithContext(ctx).Build("")
	}

	err := etcd.List(ctx, client, keyBuilderFunc, ptr.Interface(), &store.SelectionPredicate{})
	if err != nil {
		return nil, fmt.Errorf("error creating ResourceCacher: %s", err)
	}

	results := ptr.Elem()
	resources := make([]corev2.Resource, results.Len())
	for i := 0; i < results.Len(); i++ {
		r, ok := results.Index(i).Interface().(corev2.Resource)
		if !ok {
			logger.Errorf("%T is not core2.Resource", results.Index(i).Interface())
			continue
		}
		resources[i] = r
	}

	mapCache := make(map[key]Value, len(resources))
	for _, resource := range resources {
		mapCache[getCacheKey(resource)] = getCacheValue(resource, synthesize)
	}

	cache := &Resource{
		sliceCache: MakeSliceCache(resources, synthesize),
		mapCache:   mapCache,
		watcher:    etcd.GetResourceWatcher(ctx, client, keyBuilderFunc(ctx, ""), typeOfResource),
		synthesize: synthesize,
	}
	go cache.start(ctx)
	return cache, nil
}

// NewFromResources creates a new resources cache using the given resources.
// This function should only be used for testing purpose; it provides a way to
// inject resources directly into the cache without an actual store
func NewFromResources(resources []corev2.Resource, synthesize bool) *Resource {
	return &Resource{
		cacheMu:    sync.Mutex{},
		sliceCache: MakeSliceCache(resources, synthesize),
	}
}

// Watch allows cache users to get notified when the cache has new values.
// When the context is canceled, the channel will be closed.
func (r *Resource) Watch(ctx context.Context) <-chan struct{} {
	watcher := cacheWatcher{
		ctx: ctx,
		ch:  make(chan struct{}, 1),
	}
	r.watchersMu.Lock()
	r.watchers = append(r.watchers, watcher)
	r.watchersMu.Unlock()
	return watcher.ch
}

func (r *Resource) notifyWatchers() {
	r.watchersMu.Lock()
	defer r.watchersMu.Unlock()
	deletes := map[int]struct{}{}
	for i, watcher := range r.watchers {
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
	newWatchers := make([]cacheWatcher, 0, len(r.watchers))
	for i, w := range r.watchers {
		if _, ok := deletes[i]; !ok {
			newWatchers = append(newWatchers, w)
		}
	}
	r.watchers = newWatchers
}

func (r *Resource) start(ctx context.Context) {
	// 1s is the minimum scheduling interval, and so is the rate that
	// the cache will update at.
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-r.watcher:
			r.updates = append(r.updates, event)
		case <-ticker.C:
			if len(r.updates) > 0 {
				r.updateCache()
				r.notifyWatchers()
			}
		}
	}
}

// Get returns all cached resources in a namespace.
func (r *Resource) Get(namespace string) []Value {
	r.cacheMu.Lock()
	cache := r.sliceCache
	r.cacheMu.Unlock()
	start := sort.Search(len(cache), func(i int) bool {
		return cache[i].Resource.GetObjectMeta().Namespace >= namespace
	})
	endNS := namespace + string(rune(0))
	stop := sort.Search(len(cache), func(i int) bool {
		return cache[i].Resource.GetObjectMeta().Namespace >= endNS
	})
	if stop > len(cache) {
		stop = len(cache)
	}
	return cache[start:stop]
}

func (r *Resource) updateCache() {
	for _, event := range r.updates {
		resource := event.Resource
		if resource == nil {
			logger.Error("nil resource in watch event")
			continue
		}
		key := getCacheKey(resource)
		switch event.Action {
		case store.WatchCreate, store.WatchUpdate:
			r.mapCache[key] = getCacheValue(resource, r.synthesize)
		case store.WatchDelete:
			delete(r.mapCache, key)
		default:
			logger.Error("error in resource watcher")
		}
	}
	r.updates = nil
	newSliceCache := make([]Value, 0, len(r.mapCache))
	for _, v := range r.mapCache {
		newSliceCache = append(newSliceCache, v)
	}
	sort.Sort(resourceSlice(newSliceCache))
	r.cacheMu.Lock()
	r.sliceCache = newSliceCache
	r.cacheMu.Unlock()
}
