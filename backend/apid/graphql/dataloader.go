package graphql

import (
	"context"
	"errors"
	"strings"

	"github.com/graph-gophers/dataloader/v7"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
)

type key int

const (
	loadersKey key = iota

	// chunk size used by dataloader when retrieving resources from the store
	loaderPageSize = 250

	// the maximum number of records that will be read from the store by the
	// dataloader; too many puts significant strain on memory.
	maxLengthEntityDataloader  = 1_000
	maxLengthEventDataloader   = 1_000
	maxLengthGenericDataloader = 2_500
)

var errLoadersNotFound = errors.New("loaders was not found inside context")

// graphqlLoaders holds one typed loader per resource kind.
type graphqlLoaders struct {
	assets       *dataloader.Loader[string, []*corev2.Asset]
	checkConfigs *dataloader.Loader[string, []*corev2.CheckConfig]
	entities     *dataloader.Loader[string, []*corev2.Entity]
	events       *dataloader.Loader[string, []*corev2.Event]
	eventFilters *dataloader.Loader[string, []*corev2.EventFilter]
	handlers     *dataloader.Loader[string, []*corev2.Handler]
	mutators     *dataloader.Loader[string, []*corev2.Mutator]
	namespaces   *dataloader.Loader[string, []*corev2.Namespace]
	silenceds    *dataloader.Loader[string, []*corev2.Silenced]
}

// newLoader constructs a typed loader with batch capacity 1 (cache-only mode).
// When disableCache is true a NoCache implementation is used, which is useful
// in tests that need to bypass the per-request cache.
func newLoader[K comparable, V any](fn dataloader.BatchFunc[K, V], disableCache bool) *dataloader.Loader[K, V] {
	opts := []dataloader.Option[K, V]{dataloader.WithBatchCapacity[K, V](1)}
	if disableCache {
		opts = append(opts, dataloader.WithCache[K, V](&dataloader.NoCache[K, V]{}))
	}
	return dataloader.NewBatchedLoader(fn, opts...)
}

func getLoaders(ctx context.Context) (*graphqlLoaders, error) {
	loaders, ok := ctx.Value(loadersKey).(*graphqlLoaders)
	if !ok || loaders == nil {
		return nil, errLoadersNotFound
	}
	return loaders, nil
}

// assets

func loadAssetsBatchFn(c AssetClient) dataloader.BatchFunc[string, []*corev2.Asset] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[[]*corev2.Asset] {
		results := make([]*dataloader.Result[[]*corev2.Asset], 0, len(keys))
		for _, key := range keys {
			ctx := store.NamespaceContext(ctx, key)
			ctx = context.WithValue(ctx, corev2.PageSizeKey, maxLengthGenericDataloader)
			records, err := c.ListAssets(ctx)
			results = append(results, &dataloader.Result[[]*corev2.Asset]{Data: records, Error: handleListErr(err)})
		}
		return results
	}
}

func loadAssets(ctx context.Context, ns string) ([]*corev2.Asset, error) {
	loaders, err := getLoaders(ctx)
	if err != nil {
		return nil, err
	}
	return loaders.assets.Load(ctx, ns)()
}

// checks

func loadCheckConfigsBatchFn(c CheckClient) dataloader.BatchFunc[string, []*corev2.CheckConfig] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[[]*corev2.CheckConfig] {
		results := make([]*dataloader.Result[[]*corev2.CheckConfig], 0, len(keys))
		for _, key := range keys {
			ctx := store.NamespaceContext(ctx, key)
			ctx = context.WithValue(ctx, corev2.PageSizeKey, maxLengthGenericDataloader)
			records, err := c.ListChecks(ctx)
			results = append(results, &dataloader.Result[[]*corev2.CheckConfig]{Data: records, Error: handleListErr(err)})
		}
		return results
	}
}

func loadCheckConfigs(ctx context.Context, ns string) ([]*corev2.CheckConfig, error) {
	loaders, err := getLoaders(ctx)
	if err != nil {
		return nil, err
	}
	return loaders.checkConfigs.Load(ctx, ns)()
}

// entities

func listEntities(ctx context.Context, c EntityClient, maxSize int) (records []*corev2.Entity, err error) {
	pred := &store.SelectionPredicate{Continue: "", Limit: int64(loaderPageSize)}
	for {
		r, err := c.ListEntities(ctx, pred)
		if err != nil {
			return records, err
		}
		records = append(records, r...)
		if pred.Continue == "" || len(r) < loaderPageSize || len(records) >= maxSize {
			break
		}
	}
	return
}

func loadEntitiesBatchFn(c EntityClient) dataloader.BatchFunc[string, []*corev2.Entity] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[[]*corev2.Entity] {
		results := make([]*dataloader.Result[[]*corev2.Entity], 0, len(keys))
		for _, key := range keys {
			ctx := store.NamespaceContext(ctx, key)
			records, err := listEntities(ctx, c, maxLengthEntityDataloader)
			results = append(results, &dataloader.Result[[]*corev2.Entity]{Data: records, Error: handleListErr(err)})
		}
		return results
	}
}

func loadEntities(ctx context.Context, ns string) ([]*corev2.Entity, error) {
	loaders, err := getLoaders(ctx)
	if err != nil {
		return nil, err
	}
	return loaders.entities.Load(ctx, ns)()
}

// events

type eventCacheKey struct {
	namespace string
	entity    string
}

func newEventCacheKey(key string) *eventCacheKey {
	els := strings.SplitN(key, "\n", 2)
	return &eventCacheKey{namespace: els[0], entity: els[1]}
}

func (k *eventCacheKey) String() string {
	return strings.Join([]string{k.namespace, k.entity}, "\n")
}

func listEvents(ctx context.Context, c EventClient, entity string, maxSize int) ([]*corev2.Event, error) {
	pred := &store.SelectionPredicate{Continue: "", Limit: int64(loaderPageSize)}
	list := func(ctx context.Context, entity string, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
		if entity == "" {
			return c.ListEvents(ctx, pred)
		}
		return c.ListEventsByEntity(ctx, entity, pred)
	}
	results := []*corev2.Event{}
	for {
		r, err := list(ctx, entity, pred)
		if err != nil {
			return results, err
		}
		results = append(results, r...)
		if pred.Continue == "" || len(r) < loaderPageSize || len(results) >= maxSize {
			break
		}
	}
	return results, nil
}

func loadEventsBatchFn(c EventClient) dataloader.BatchFunc[string, []*corev2.Event] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[[]*corev2.Event] {
		results := make([]*dataloader.Result[[]*corev2.Event], 0, len(keys))
		for _, key := range keys {
			eck := newEventCacheKey(key)
			ctx := store.NamespaceContext(ctx, eck.namespace)
			records, err := listEvents(ctx, c, eck.entity, maxLengthEventDataloader)
			results = append(results, &dataloader.Result[[]*corev2.Event]{Data: records, Error: handleListErr(err)})
		}
		return results
	}
}

func loadEvents(ctx context.Context, ns, entity string) ([]*corev2.Event, error) {
	loaders, err := getLoaders(ctx)
	if err != nil {
		return nil, err
	}
	key := &eventCacheKey{namespace: ns, entity: entity}
	return loaders.events.Load(ctx, key.String())()
}

// event filters

func loadEventFiltersBatchFn(c EventFilterClient) dataloader.BatchFunc[string, []*corev2.EventFilter] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[[]*corev2.EventFilter] {
		results := make([]*dataloader.Result[[]*corev2.EventFilter], 0, len(keys))
		for _, key := range keys {
			ctx := store.NamespaceContext(ctx, key)
			ctx = context.WithValue(ctx, corev2.PageSizeKey, maxLengthGenericDataloader)
			records, err := c.ListEventFilters(ctx)
			results = append(results, &dataloader.Result[[]*corev2.EventFilter]{Data: records, Error: handleListErr(err)})
		}
		return results
	}
}

func loadEventFilters(ctx context.Context, ns string) ([]*corev2.EventFilter, error) {
	loaders, err := getLoaders(ctx)
	if err != nil {
		return nil, err
	}
	return loaders.eventFilters.Load(ctx, ns)()
}

// handlers

func loadHandlersBatchFn(c HandlerClient) dataloader.BatchFunc[string, []*corev2.Handler] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[[]*corev2.Handler] {
		results := make([]*dataloader.Result[[]*corev2.Handler], 0, len(keys))
		for _, key := range keys {
			ctx := store.NamespaceContext(ctx, key)
			ctx = context.WithValue(ctx, corev2.PageSizeKey, maxLengthGenericDataloader)
			records, err := c.ListHandlers(ctx)
			results = append(results, &dataloader.Result[[]*corev2.Handler]{Data: records, Error: handleListErr(err)})
		}
		return results
	}
}

func loadHandlers(ctx context.Context, ns string) ([]*corev2.Handler, error) {
	loaders, err := getLoaders(ctx)
	if err != nil {
		return nil, err
	}
	return loaders.handlers.Load(ctx, ns)()
}

// mutators

func loadMutatorsBatchFn(c MutatorClient) dataloader.BatchFunc[string, []*corev2.Mutator] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[[]*corev2.Mutator] {
		results := make([]*dataloader.Result[[]*corev2.Mutator], 0, len(keys))
		for _, key := range keys {
			ctx := store.NamespaceContext(ctx, key)
			ctx = context.WithValue(ctx, corev2.PageSizeKey, maxLengthGenericDataloader)
			records, err := c.ListMutators(ctx)
			results = append(results, &dataloader.Result[[]*corev2.Mutator]{Data: records, Error: handleListErr(err)})
		}
		return results
	}
}

func loadMutators(ctx context.Context, ns string) ([]*corev2.Mutator, error) {
	loaders, err := getLoaders(ctx)
	if err != nil {
		return nil, err
	}
	return loaders.mutators.Load(ctx, ns)()
}

// namespaces

func loadNamespacesBatchFn(c NamespaceClient) dataloader.BatchFunc[string, []*corev2.Namespace] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[[]*corev2.Namespace] {
		results := make([]*dataloader.Result[[]*corev2.Namespace], 0, len(keys))
		for range keys {
			ctx := context.WithValue(ctx, corev2.PageSizeKey, maxLengthGenericDataloader)
			records, err := c.ListNamespaces(ctx, &store.SelectionPredicate{})
			results = append(results, &dataloader.Result[[]*corev2.Namespace]{Data: records, Error: handleListErr(err)})
		}
		return results
	}
}

func loadNamespaces(ctx context.Context) ([]*corev2.Namespace, error) {
	loaders, err := getLoaders(ctx)
	if err != nil {
		return nil, err
	}
	return loaders.namespaces.Load(ctx, "*")()
}

// silences

func loadSilencedsBatchFn(c SilencedClient) dataloader.BatchFunc[string, []*corev2.Silenced] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[[]*corev2.Silenced] {
		results := make([]*dataloader.Result[[]*corev2.Silenced], 0, len(keys))
		for _, key := range keys {
			ctx := store.NamespaceContext(ctx, key)
			ctx = context.WithValue(ctx, corev2.PageSizeKey, maxLengthGenericDataloader)
			records, err := c.ListSilenced(ctx)
			results = append(results, &dataloader.Result[[]*corev2.Silenced]{Data: records, Error: handleListErr(err)})
		}
		return results
	}
}

func loadSilenceds(ctx context.Context, ns string) ([]*corev2.Silenced, error) {
	loaders, err := getLoaders(ctx)
	if err != nil {
		return nil, err
	}
	return loaders.silenceds.Load(ctx, ns)()
}

func contextWithLoaders(ctx context.Context, cfg ServiceConfig, disableCache bool) context.Context {
	loaders := &graphqlLoaders{
		assets:       newLoader(loadAssetsBatchFn(cfg.AssetClient), disableCache),
		checkConfigs: newLoader(loadCheckConfigsBatchFn(cfg.CheckClient), disableCache),
		entities:     newLoader(loadEntitiesBatchFn(cfg.EntityClient), disableCache),
		events:       newLoader(loadEventsBatchFn(cfg.EventClient), disableCache),
		eventFilters: newLoader(loadEventFiltersBatchFn(cfg.EventFilterClient), disableCache),
		handlers:     newLoader(loadHandlersBatchFn(cfg.HandlerClient), disableCache),
		mutators:     newLoader(loadMutatorsBatchFn(cfg.MutatorClient), disableCache),
		namespaces:   newLoader(loadNamespacesBatchFn(cfg.NamespaceClient), disableCache),
		silenceds:    newLoader(loadSilencedsBatchFn(cfg.SilencedClient), disableCache),
	}
	return context.WithValue(ctx, loadersKey, loaders)
}

// When resolving a field, GraphQL does not consider the absence of a value an
// error; as such we omit the error if the API client returns Permission denied.
func handleListErr(err error) error {
	if err == authorization.ErrUnauthorized || err == authorization.ErrNoClaims {
		logger.WithError(err).Warn("couldn't access resource")
		return nil
	}
	return err
}

// When resolving a field, GraphQL does not consider the absence of a value an
// error; as such we omit the error when the API client returns NotFound or
// Permission denied.
func handleFetchResult(resource interface{}, err error) (interface{}, error) {
	if err == authorization.ErrUnauthorized || err == authorization.ErrNoClaims {
		logger.WithError(err).Warn("couldn't access resource")
		return nil, nil
	}
	if _, ok := err.(*store.ErrNotFound); ok {
		logger.WithError(err).Warn("couldn't access resource")
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return resource, err
}
