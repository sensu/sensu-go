package cache

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/types/dynamic"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Cache interfaces the cache.Resource struct for easier testing
type Cache interface {
	Get(namespace string) []Value
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

func (s resourceSlice) Find(value Value) Value {
	idx := sort.Search(len(s), func(i int) bool {
		return !resourceLT(s[i], value)
	})
	if idx < len(s) && s[idx].Resource.GetObjectMeta().Name == value.Resource.GetObjectMeta().Name {
		return s[idx]
	}
	return Value{}
}

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
// `sync/atomic` expects the first word in an allocated struct to be 64-bit
// aligned on both ARM and x86-32. See https://goo.gl/zW7dgq for more details.
type Resource struct {
	count      int64
	cache      cache
	cacheMu    sync.RWMutex
	watchers   []cacheWatcher
	watchersMu sync.Mutex
	synthesize bool
	resourceT  corev2.Resource
	client     *clientv3.Client
}

// getResources retrieves the resources from the store
func getResources(ctx context.Context, client *clientv3.Client, resource corev2.Resource) ([]corev2.Resource, error) {
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
	return resources, nil
}

// New creates a new resource cache. It retrieves all resources from the
// store on creation.
func New(ctx context.Context, client *clientv3.Client, resource corev2.Resource, synthesize bool) (*Resource, error) {
	resources, err := getResources(ctx, client, resource)
	if err != nil {
		return nil, err
	}

	cache := buildCache(resources, synthesize)

	cacher := &Resource{
		cache:      cache,
		synthesize: synthesize,
		resourceT:  resource,
		client:     client,
	}
	atomic.StoreInt64(&cacher.count, int64(len(resources)))

	go cacher.start(ctx)

	return cacher, nil
}

// NewFromResources creates a new resources cache using the given resources.
// This function should only be used for testing purpose; it provides a way to
// inject resources directly into the cache without an actual store
func NewFromResources(resources []corev2.Resource, synthesize bool) *Resource {
	return &Resource{
		cacheMu: sync.RWMutex{},
		cache:   buildCache(resources, synthesize),
	}
}

// Get returns all cached resources in a namespace.
func (r *Resource) Get(namespace string) []Value {
	r.cacheMu.RLock()
	defer r.cacheMu.RUnlock()
	return r.cache[namespace]
}

// GetAll returns all cached resources across all namespaces.
func (r *Resource) GetAll() []Value {
	values := []Value{}
	r.cacheMu.RLock()
	defer r.cacheMu.RUnlock()
	for _, n := range r.cache {
		values = append(values, n...)
	}
	return values
}

// Count returns the total count of all cached resources across all namespaces.
func (r *Resource) Count() int64 {
	return atomic.LoadInt64(&r.count)
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
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			updates, err := r.rebuild(ctx)
			if err != nil {
				logger.WithError(err).Error("couldn't rebuild cache")
			}
			if updates {
				r.notifyWatchers()
			}
		}
	}
}

// rebuild the cache using the store as the source of truth
func (r *Resource) rebuild(ctx context.Context) (bool, error) {
	logger.Debugf("rebuilding the cache for resource type %T", r.resourceT)
	resources, err := getResources(ctx, r.client, r.resourceT)
	if err != nil {
		return false, err
	}
	atomic.StoreInt64(&r.count, int64(len(resources)))
	newCache := buildCache(resources, r.synthesize)
	var hasUpdates bool
	r.cacheMu.Lock()
	defer r.cacheMu.Unlock()
	for key, values := range newCache {
		oldValues, ok := r.cache[key]
		if !ok {
			// Apparently we have an entire namespace's worth of values that
			// just appeared out of nowhere...
			hasUpdates = true
			continue
		}
		for _, value := range oldValues {
			newValue := resourceSlice(values).Find(value)
			if newValue.Resource == nil {
				hasUpdates = true
				continue
			}
		}
		for _, value := range values {
			oldValue := resourceSlice(oldValues).Find(value)
			if oldValue.Resource == nil {
				hasUpdates = true
				continue
			}
			if !reflect.DeepEqual(oldValue.Resource, value.Resource) {
				hasUpdates = true
				continue
			}
		}
	}
	r.cache = newCache
	return hasUpdates, nil
}
