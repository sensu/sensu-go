package v2

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/dynamic"
)

// Value contains a cached value, and its synthesized companion.
type Value[R storev2.Resource[T], T any] struct {
	Resource R
	Synth    interface{}
}

func getCacheKey(resource corev3.Resource) string {
	return resource.GetMetadata().Namespace
}

func getCacheValue[R storev2.Resource[T], T any](resource R, synthesize bool) Value[R, T] {
	v := Value[R, T]{Resource: resource}
	if synthesize {
		v.Synth = dynamic.Synthesize(resource)
	}
	return v
}

type cache[R storev2.Resource[T], T any] map[string][]Value[R, T]

func buildCache[R storev2.Resource[T], T any](resources []R, synthesize bool) (cache[R, T], string) {
	cache := make(map[string][]Value[R, T])
	var newestUpdate time.Time
	for i, resource := range resources {
		meta := resource.GetMetadata()
		var updatedSince time.Time
		if err := updatedSince.UnmarshalText([]byte(meta.Labels[store.SensuUpdatedAtKey])); err == nil {
			if updatedSince.After(newestUpdate) {
				newestUpdate = updatedSince
			}
		}
		key := getCacheKey(resource)
		cache[key] = append(cache[key], getCacheValue[R, T](resources[i], synthesize))
	}
	for _, slice := range cache {
		sort.Sort(resourceSlice[R, T](slice))
	}
	b, _ := newestUpdate.MarshalText()
	return cache, string(b)
}

type resourceSlice[R storev2.Resource[T], T any] []Value[R, T]

func (s resourceSlice[R, T]) Find(value Value[R, T]) Value[R, T] {
	idx := sort.Search(len(s), func(i int) bool {
		return !resourceLT(s[i], value)
	})
	if idx < len(s) && s[idx].Resource.GetMetadata().Name == value.Resource.GetMetadata().Name {
		return s[idx]
	}
	return Value[R, T]{}
}

func (s resourceSlice[R, T]) FindIndex(name string) (int, bool) {
	idx := sort.Search(len(s), func(i int) bool {
		return s[i].Resource.GetMetadata().Name >= name
	})
	if idx < len(s) && s[idx].Resource.GetMetadata().Name == name {
		return idx, true
	}
	return idx, false
}

func (s resourceSlice[R, T]) Len() int {
	return len(s)
}

func (s resourceSlice[R, T]) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s resourceSlice[R, T]) Less(i, j int) bool {
	return resourceLT(s[i], s[j])
}

func resourceLT[R storev2.Resource[T], T any](x, y Value[R, T]) bool {
	if x.Resource == nil {
		return true
	}
	if y.Resource == nil {
		return false
	}
	return x.Resource.GetMetadata().Name < y.Resource.GetMetadata().Name
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
type Resource[R storev2.Resource[T], T any] struct {
	count        int64
	cache        cache[R, T]
	cacheMu      sync.RWMutex
	watchers     []cacheWatcher
	watchersMu   sync.Mutex
	synthesize   bool
	store        storev2.Interface
	updatedSince string
}

// New creates a new resource cache. It retrieves all resources from the
// store on creation.
func New[R storev2.Resource[T], T any](ctx context.Context, store storev2.Interface, synthesize bool) (*Resource[R, T], error) {
	gstore := storev2.Of[R, T](store)
	resources, err := gstore.List(ctx, storev2.ID{}, nil)
	if err != nil {
		return nil, err
	}

	cache, updatedSince := buildCache[R, T](resources, synthesize)

	cacher := &Resource[R, T]{
		cache:        cache,
		synthesize:   synthesize,
		store:        store,
		updatedSince: updatedSince,
	}
	atomic.StoreInt64(&cacher.count, int64(len(resources)))

	go cacher.start(ctx)

	return cacher, nil
}

// NewFromResources creates a new resources cache using the given resources.
// This function should only be used for testing purpose; it provides a way to
// inject resources directly into the cache without an actual store
func NewFromResources[R storev2.Resource[T], T any](resources []R, synthesize bool) *Resource[R, T] {
	cache, updatedSince := buildCache[R, T](resources, synthesize)
	return &Resource[R, T]{
		cacheMu:      sync.RWMutex{},
		cache:        cache,
		updatedSince: updatedSince,
	}
}

// Get returns all cached resources in a namespace.
func (r *Resource[R, T]) Get(namespace string) []Value[R, T] {
	r.cacheMu.RLock()
	defer r.cacheMu.RUnlock()
	return r.cache[namespace]
}

// GetAll returns all cached resources across all namespaces.
func (r *Resource[R, T]) GetAll() []Value[R, T] {
	values := []Value[R, T]{}
	r.cacheMu.RLock()
	defer r.cacheMu.RUnlock()
	for _, n := range r.cache {
		values = append(values, n...)
	}
	return values
}

// Count returns the total count of all cached resources across all namespaces.
func (r *Resource[R, T]) Count() int64 {
	return atomic.LoadInt64(&r.count)
}

// Watch allows cache users to get notified when the cache has new values.
// When the context is canceled, the channel will be closed.
func (r *Resource[R, T]) Watch(ctx context.Context) <-chan struct{} {
	watcher := cacheWatcher{
		ctx: ctx,
		ch:  make(chan struct{}, 1),
	}
	r.watchersMu.Lock()
	r.watchers = append(r.watchers, watcher)
	r.watchersMu.Unlock()
	return watcher.ch
}

func (r *Resource[R, T]) notifyWatchers() {
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

func (r *Resource[R, T]) start(ctx context.Context) {
	// 1s is the minimum scheduling interval, and so is the rate that
	// the cache will update at.
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			updates, err := r.refresh(ctx)
			if err != nil {
				logger.WithError(err).Error("couldn't refresh cache")
			}
			if updates {
				r.notifyWatchers()
			}
		}
	}
}

// refresh the cache using the store as the source of truth
func (r *Resource[R, T]) refresh(ctx context.Context) (bool, error) {
	gstore := storev2.Of[R, T](r.store)
	pred := &store.SelectionPredicate{
		UpdatedSince:   r.updatedSince,
		IncludeDeletes: true,
	}
	resources, err := gstore.List(ctx, storev2.ID{}, pred)
	if err != nil {
		return false, err
	}
	r.cacheMu.Lock()
	defer r.cacheMu.Unlock()
	hasUpdates := len(resources) > 0
	var newestUpdate time.Time
	_ = newestUpdate.UnmarshalText([]byte(r.updatedSince))
	for _, resource := range resources {
		meta := resource.GetMetadata()
		var updatedAt time.Time
		if err := updatedAt.UnmarshalText([]byte(meta.Labels[store.SensuUpdatedAtKey])); err == nil {
			if updatedAt.After(newestUpdate) {
				newestUpdate = updatedAt
			}
		}
		slice := resourceSlice[R, T](r.cache[meta.Namespace])
		if _, ok := meta.Labels[store.SensuDeletedAtKey]; ok {
			toDelete, ok := slice.FindIndex(meta.Name)
			if ok {
				// remove value
				slice = append(slice[:toDelete], slice[toDelete+1:]...)
				r.cache[meta.Namespace] = slice
			}
		} else {
			value := getCacheValue(resource, r.synthesize)
			idx, ok := slice.FindIndex(meta.Name)
			if ok {
				// replace value
				slice[idx] = value
			} else {
				// new value, sorted insert
				slice = append(slice, value)
				copy(slice[idx+1:], slice[idx:])
				slice[idx] = value
				r.cache[meta.Namespace] = slice
			}
		}
	}
	var resourceCount int64
	for _, slice := range r.cache {
		resourceCount += int64(len(slice))
	}
	atomic.StoreInt64(&r.count, resourceCount)
	b, _ := newestUpdate.MarshalText()
	r.updatedSince = string(b)

	return hasUpdates, nil
}
