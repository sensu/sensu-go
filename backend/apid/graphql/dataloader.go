package graphql

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/graph-gophers/dataloader"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

type key int

const (
	loadersKey key = iota
	entitiesLoaderKey
	eventsLoaderKey
	genericLoaderKey
	namespacesLoaderKey

	// chunk size used by dataloader when retrieving resources from the store
	loaderPageSize = 250

	// the maximum number of records that will be read from the store by the
	// dataloader; too many can put significant strain on memory.
	maxLengthEntityDataloader  = 1_000
	maxLengthEventDataloader   = 1_000
	maxLengthGenericDataloader = 5_000
)

var (
	errLoadersNotFound        = errors.New("loaders was not found inside context")
	errLoaderNotFound         = errors.New("loader was not found")
	errUnexpectedLoaderResult = errors.New("loader returned unexpected result")
)

// entities

func listEntities(ctx context.Context, c EntityClient, maxSize int) ([]*corev2.Entity, error) {
	pred := &store.SelectionPredicate{Continue: "", Limit: int64(loaderPageSize)}
	results := []*corev2.Entity{}
	for {
		r, err := c.ListEntities(ctx, pred)
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

func loadEntitiesBatchFn(c EntityClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			ctx := store.NamespaceContext(ctx, key.String())
			records, err := listEntities(ctx, c, 1000)
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadEntities(ctx context.Context, ns string) ([]*corev2.Entity, error) {
	var records []*corev2.Entity
	loader, err := getLoader(ctx, entitiesLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey(ns))()
	records, ok := results.([]*corev2.Entity)
	if err == nil && !ok {
		err = fmt.Errorf("entity loader: %s", errUnexpectedLoaderResult)
	}
	return records, err
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

func (k *eventCacheKey) Raw() interface{} {
	return k
}

func listEvents(ctx context.Context, c EventClient, entity string) ([]*corev2.Event, error) {
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
		if pred.Continue == "" || len(r) < loaderPageSize {
			break
		}
	}
	return results, nil
}

func loadEventsBatchFn(c EventClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			key := newEventCacheKey(key.String())
			ctx := store.NamespaceContext(ctx, key.namespace)
			records, err := listEvents(ctx, c, key.entity)
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadEvents(ctx context.Context, ns, entity string) ([]*corev2.Event, error) {
	var records []*corev2.Event
	loader, err := getLoader(ctx, eventsLoaderKey)
	if err != nil {
		return records, err
	}

	key := &eventCacheKey{namespace: ns, entity: entity}
	results, err := loader.Load(ctx, key)()
	records, ok := results.([]*corev2.Event)
	if err == nil && !ok {
		err = fmt.Errorf("event loader: %s", errUnexpectedLoaderResult)
	}
	return records, err
}

// generic loader

type loadResourceReq struct {
	namespace string
	apigroup  string
	typename  string
}

func hydrateLoadResourceReq(key string) *loadResourceReq {
	els := strings.SplitN(key, "\n", 2)
	return &loadResourceReq{apigroup: els[0], typename: els[1], namespace: els[2]}
}

func (k *loadResourceReq) String() string {
	return strings.Join([]string{k.apigroup, k.typename, k.namespace}, "\n")
}

func (k *loadResourceReq) Raw() interface{} {
	return k
}

// sliceFromTypeMeta returns a slice of the type corresponding to the given
// api_group and type name.
func sliceFromTypeMeta(apigroup, typename string) (slice interface{}, err error) {
	t, err := types.ResolveType(apigroup, typename)
	if err != nil {
		return []interface{}{}, err
	}
	objT := reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(reflect.TypeOf(t).Elem())), 0, 0)
	objs := reflect.New(objT.Type())
	objs.Elem().Set(objT)
	return objs.Interface(), nil
}

func listResource(ctx context.Context, c GenericClient, resources interface{}, maxSize int) error {
	pred := &store.SelectionPredicate{Continue: "", Limit: int64(loaderPageSize)}
	results := reflect.ValueOf(resources).Elem()
	for {
		res := reflect.MakeSlice(reflect.TypeOf(resources), 0, 0)
		err := c.List(ctx, res.Interface(), pred)
		if err != nil {
			return err
		}
		results.Set(reflect.AppendSlice(results, res))
		if pred.Continue == "" || res.Len() < loaderPageSize || results.Len() >= maxSize {
			break
		}
	}
	return nil
}

func loadResourceBatchFn(c GenericClient, maxSize int) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			key := hydrateLoadResourceReq(key.String())
			ctx := store.NamespaceContext(ctx, key.namespace)
			records, err := sliceFromTypeMeta(key.apigroup, key.typename)
			if err == nil {
				err = listResource(ctx, c, records, maxSize)
			}
			result := &dataloader.Result{Data: records, Error: err}
			results = append(results, result)
		}
		return results
	}
}

func loadResource(ctx context.Context, req *loadResourceReq, res interface{}) error {
	loaded := reflect.ValueOf(res).Elem()
	loader, err := getLoader(ctx, genericLoaderKey)
	if err != nil {
		return err
	}
	result, err := loader.Load(ctx, req)()
	if err != nil {
		return err
	}
	loaded.Set(reflect.ValueOf(result))
	return nil
}

// namespaces

func loadNamespacesBatchFn(c NamespaceClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for range keys {
			records, err := c.ListNamespaces(ctx, &store.SelectionPredicate{})
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadNamespaces(ctx context.Context) ([]*corev2.Namespace, error) {
	var records []*corev2.Namespace
	loader, err := getLoader(ctx, namespacesLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey("*"))()
	records, ok := results.([]*corev2.Namespace)
	if err == nil && !ok {
		err = fmt.Errorf("namespace loader: %s", errUnexpectedLoaderResult)
	}
	return records, err
}

// helpers

func loadSilenceds(ctx context.Context, namespace string) ([]*corev2.Silenced, error) {
	results := []*corev2.Silenced{}
	req := loadResourceReq{namespace: namespace, typename: "Silenced", apigroup: "core/v2"}
	err := loadResource(ctx, &req, results)
	return results, err
}

func loadHandlers(ctx context.Context, namespace string) ([]*corev2.Handler, error) {
	results := []*corev2.Handler{}
	req := loadResourceReq{namespace: namespace, typename: "Handler", apigroup: "core/v2"}
	err := loadResource(ctx, &req, results)
	return results, err
}

func loadAssets(ctx context.Context, namespace string) ([]*corev2.Asset, error) {
	results := []*corev2.Asset{}
	req := loadResourceReq{namespace: namespace, typename: "Asset", apigroup: "core/v2"}
	err := loadResource(ctx, &req, results)
	return results, err
}

func contextWithLoaders(ctx context.Context, cfg ServiceConfig, opts ...dataloader.Option) context.Context {
	// Currently all fields are resolved serially, as such we disable batching and
	// rely only on dataloader's cache.
	opts = append([]dataloader.Option{dataloader.WithBatchCapacity(1)}, opts...)

	loaders := map[key]*dataloader.Loader{}
	loaders[entitiesLoaderKey] = dataloader.NewBatchedLoader(loadEntitiesBatchFn(cfg.EntityClient), opts...)
	loaders[eventsLoaderKey] = dataloader.NewBatchedLoader(loadEventsBatchFn(cfg.EventClient), opts...)
	loaders[genericLoaderKey] = dataloader.NewBatchedLoader(loadResourceBatchFn(cfg.GenericClient, maxLengthGenericDataloader), opts...)
	loaders[namespacesLoaderKey] = dataloader.NewBatchedLoader(loadNamespacesBatchFn(cfg.NamespaceClient), opts...)
	return context.WithValue(ctx, loadersKey, loaders)
}

func getLoader(ctx context.Context, loaderKey key) (*dataloader.Loader, error) {
	loaders, ok := ctx.Value(loadersKey).(map[key]*dataloader.Loader)
	if !ok {
		return nil, errLoadersNotFound
	}

	loader, ok := loaders[loaderKey]
	if !ok {
		return loader, errLoaderNotFound
	}
	return loader, nil
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
