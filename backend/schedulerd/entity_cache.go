package schedulerd

import (
	"context"
	"fmt"
	"sort"
	"sync"

	time "github.com/echlebek/timeproxy"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types/dynamic"
)

type entityCacheKey struct {
	Name      string
	Namespace string
}

// EntityCacheValue contains an entity, and its synthesized companion.
type EntityCacheValue struct {
	Entity *corev2.Entity
	Synth  interface{}
}

func getEntityCacheKey(entity *corev2.Entity) entityCacheKey {
	return entityCacheKey{
		Namespace: entity.Namespace,
		Name:      entity.Name,
	}
}

func getEntityCacheValue(entity *corev2.Entity) EntityCacheValue {
	return EntityCacheValue{
		Entity: entity,
		Synth:  dynamic.Synthesize(entity),
	}
}

func makeSliceCache(entities []*corev2.Entity) []EntityCacheValue {
	cache := make([]EntityCacheValue, len(entities))
	for i := range cache {
		cache[i] = getEntityCacheValue(entities[i])
	}
	return cache
}

type entitySlice []EntityCacheValue

func (e entitySlice) Len() int {
	return len(e)
}

func (e entitySlice) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func entityLT(x, y EntityCacheValue) bool {
	if x.Entity == nil {
		return true
	}
	if y.Entity == nil {
		return false
	}
	if x.Entity.Namespace == y.Entity.Namespace {
		return x.Entity.Name < y.Entity.Name
	}
	return x.Entity.Namespace < y.Entity.Namespace
}

func (e entitySlice) Less(i, j int) bool {
	return entityLT(e[i], e[j])
}

type cacheWatcher struct {
	ctx context.Context
	ch  chan struct{}
}

// EntityCache is a cache of entities. The cache uses a watcher on entities in
// order to keep itself up to date. Entities can be efficiently retrieved from
// the cache by namespace.
type EntityCache struct {
	watcher    <-chan store.WatchEventEntity
	mapCache   map[entityCacheKey]EntityCacheValue
	sliceCache []EntityCacheValue
	updates    []store.WatchEventEntity
	cacheMu    sync.Mutex
	watchers   []cacheWatcher
	watchersMu sync.Mutex
}

// NewEntityCache creates a new EntityCache. It retrieves all entities from the
// store on creation.
func NewEntityCache(ctx context.Context, s store.EntityStore) (*EntityCache, error) {
	entities, err := s.GetEntities(ctx, &store.SelectionPredicate{})
	if err != nil {
		return nil, fmt.Errorf("error creating EntityCache: %s", err)
	}
	mapCache := make(map[entityCacheKey]EntityCacheValue, len(entities))
	for _, entity := range entities {
		mapCache[getEntityCacheKey(entity)] = getEntityCacheValue(entity)
	}
	cache := &EntityCache{
		sliceCache: makeSliceCache(entities),
		mapCache:   mapCache,
		watcher:    s.GetEntityWatcher(ctx),
	}
	go cache.start(ctx)
	return cache, nil
}

// Watch allows cache users to get notified when the cache has new values.
// When the context is canceled, the channel will be closed.
func (e *EntityCache) Watch(ctx context.Context) <-chan struct{} {
	watcher := cacheWatcher{
		ctx: ctx,
		ch:  make(chan struct{}, 1),
	}
	e.watchersMu.Lock()
	e.watchers = append(e.watchers, watcher)
	e.watchersMu.Unlock()
	return watcher.ch
}

func (e *EntityCache) notifyWatchers() {
	e.watchersMu.Lock()
	defer e.watchersMu.Unlock()
	deletes := map[int]struct{}{}
	for i, watcher := range e.watchers {
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
	newWatchers := make([]cacheWatcher, 0, len(e.watchers))
	for i, w := range e.watchers {
		if _, ok := deletes[i]; !ok {
			newWatchers = append(newWatchers, w)
		}
	}
	e.watchers = newWatchers
}

func (e *EntityCache) start(ctx context.Context) {
	// 1s is the minimum scheduling interval, and so is the rate that
	// the cache will update at.
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-e.watcher:
			e.updates = append(e.updates, event)
		case <-ticker.C:
			if len(e.updates) > 0 {
				e.updateCache()
				e.notifyWatchers()
			}
		}
	}
}

// GetEntities gets all entities in a namespace.
func (e *EntityCache) GetEntities(namespace string) []EntityCacheValue {
	e.cacheMu.Lock()
	cache := e.sliceCache
	e.cacheMu.Unlock()
	start := sort.Search(len(cache), func(i int) bool {
		return cache[i].Entity.Namespace >= namespace
	})
	endNS := namespace + string(rune(0))
	stop := sort.Search(len(cache), func(i int) bool {
		return cache[i].Entity.Namespace >= endNS
	})
	if stop > len(cache) {
		stop = len(cache)
	}
	return cache[start:stop]
}

func (e *EntityCache) updateCache() {
	for _, event := range e.updates {
		entity := event.Entity
		if entity == nil {
			logger.Error("nil entity in watch event")
			continue
		}
		key := getEntityCacheKey(entity)
		switch event.Action {
		case store.WatchCreate, store.WatchUpdate:
			e.mapCache[key] = getEntityCacheValue(entity)
		case store.WatchDelete:
			delete(e.mapCache, key)
		default:
			logger.Error("error in entity watcher")
		}
	}
	e.updates = nil
	newSliceCache := make([]EntityCacheValue, 0, len(e.mapCache))
	for _, v := range e.mapCache {
		newSliceCache = append(newSliceCache, v)
	}
	sort.Sort(entitySlice(newSliceCache))
	e.cacheMu.Lock()
	e.sliceCache = newSliceCache
	e.cacheMu.Unlock()
}
