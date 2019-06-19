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

func getCacheKey(resource corev2.Resource) string {
	return resource.GetObjectMeta().Namespace
}

func getCacheValue(resource corev2.Resource, synthesize bool) Value {
	v := Value{Resource: resource}
	if synthesize {
		v.Synth = dynamic.Synthesize(resource)
	}
	return v
}

type cache map[string][]Value

// buildCache ...
func buildCache(resources []corev2.Resource, synthesize bool) cache {
	cache := make(map[string][]Value)
	for i, resource := range resources {
		key := getCacheKey(resource)
		cache[key] = append(cache[key], getCacheValue(resources[i], synthesize))
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
	return x.Resource.GetObjectMeta().Name < y.Resource.GetObjectMeta().Name
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
	cache      cache
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

	cache := buildCache(resources, synthesize)

	cacher := &Resource{
		cache:      cache,
		watcher:    etcd.GetResourceWatcher(ctx, client, keyBuilderFunc(ctx, ""), typeOfResource),
		synthesize: synthesize,
	}
	go cacher.start(ctx)

	return cacher, nil
}

// NewFromResources creates a new resources cache using the given resources.
// This function should only be used for testing purpose; it provides a way to
// inject resources directly into the cache without an actual store
func NewFromResources(resources []corev2.Resource, synthesize bool) *Resource {
	return &Resource{
		cacheMu: sync.Mutex{},
		cache:   buildCache(resources, synthesize),
	}
}

// Get returns all cached resources in a namespace.
func (r *Resource) Get(namespace string) []Value {
	r.cacheMu.Lock()
	defer r.cacheMu.Unlock()
	return r.cache[namespace]
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

func (r *Resource) updateCache() {
	r.cacheMu.Lock()
	for _, event := range r.updates {
		resource := event.Resource
		if resource == nil || reflect.ValueOf(resource).IsNil() {
			logger.Error("nil resource in watch event")
			continue
		}
		key := getCacheKey(resource)

		switch event.Action {
		case store.WatchCreate:
			// Append the new resource to the corresponding namespace
			r.cache[key] = append(r.cache[key], getCacheValue(resource, r.synthesize))
		case store.WatchUpdate:
			// Loop through the resources of the resource's namespace to find the
			// exact resource and update it
			for i := range r.cache[key] {
				if r.cache[key][i].Resource.GetObjectMeta().Name == resource.GetObjectMeta().Name {
					r.cache[key][i] = getCacheValue(resource, r.synthesize)
					break
				}
			}
		case store.WatchDelete:
			// Loop through the resources of the resource's namespace to find the
			// exact resource and delete it
			for i := range r.cache[key] {
				if r.cache[key][i].Resource.GetObjectMeta().Name == resource.GetObjectMeta().Name {
					r.cache[key] = append(r.cache[key][:i], r.cache[key][i+1:]...)
					break
				}
			}
		default:
			logger.Error("error in resource watcher")
		}
	}

	r.updates = nil

	// Sort the resources alphabetically in each namespace
	for _, v := range r.cache {
		sort.Sort(resourceSlice(v))
	}

	r.cacheMu.Unlock()
}
