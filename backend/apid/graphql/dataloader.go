package graphql

import (
	"context"
	"errors"

	"github.com/graph-gophers/dataloader"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/cli/client"
)

type key int

const (
	loadersKey key = iota
	assetsLoaderKey
	checkConfigsLoaderKey
	entitiesLoaderKey
	eventsLoaderKey
	eventFiltersLoaderKey
	handlersLoaderKey
	mutatorsLoaderKey
	namespacesLoaderKey
	silencedsLoaderKey
)

var (
	errLoadersNotFound        = errors.New("loaders was not found inside context")
	errLoaderNotFound         = errors.New("loader was not found")
	errUnexpectedLoaderResult = errors.New("loader returned unexpected result")
)

// assets

func loadAssetsBatchFn(c AssetClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			ctx = store.NamespaceContext(ctx, key.String())
			records, err := c.ListAssets(ctx)
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadAssets(ctx context.Context, ns string) ([]*corev2.Asset, error) {
	var records []*corev2.Asset
	loader, err := getLoader(ctx, assetsLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey(ns))()
	records, ok := results.([]*corev2.Asset)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

// checks

func loadCheckConfigsBatchFn(c client.APIClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			records, err := c.ListChecks(key.String(), &client.ListOptions{})
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadCheckConfigs(ctx context.Context, ns string) ([]corev2.CheckConfig, error) {
	var records []corev2.CheckConfig
	loader, err := getLoader(ctx, checkConfigsLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey(ns))()
	records, ok := results.([]corev2.CheckConfig)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

// entities

func loadEntitiesBatchFn(c client.APIClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			records, err := c.ListEntities(key.String(), &client.ListOptions{})
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadEntities(ctx context.Context, ns string) ([]corev2.Entity, error) {
	var records []corev2.Entity
	loader, err := getLoader(ctx, entitiesLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey(ns))()
	records, ok := results.([]corev2.Entity)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

// events

func loadEventsBatchFn(c client.APIClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			records, err := c.ListEvents(key.String(), &client.ListOptions{})
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadEvents(ctx context.Context, ns string) ([]corev2.Event, error) {
	var records []corev2.Event
	loader, err := getLoader(ctx, eventsLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey(ns))()
	records, ok := results.([]corev2.Event)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

// event filters

func loadEventFiltersBatchFn(c client.APIClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			records, err := c.ListFilters(key.String(), &client.ListOptions{})
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadEventFilters(ctx context.Context, ns string) ([]corev2.EventFilter, error) {
	var records []corev2.EventFilter
	loader, err := getLoader(ctx, eventFiltersLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey(ns))()
	records, ok := results.([]corev2.EventFilter)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

// handlers

func loadHandlersBatchFn(c client.APIClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			records, err := c.ListHandlers(key.String(), &client.ListOptions{})
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadHandlers(ctx context.Context, ns string) ([]corev2.Handler, error) {
	var records []corev2.Handler
	loader, err := getLoader(ctx, handlersLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey(ns))()
	records, ok := results.([]corev2.Handler)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

// mutators

func loadMutatorsBatchFn(c client.APIClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			records, err := c.ListMutators(key.String(), &client.ListOptions{})
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadMutators(ctx context.Context, ns string) ([]corev2.Mutator, error) {
	var records []corev2.Mutator
	loader, err := getLoader(ctx, mutatorsLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey(ns))()
	records, ok := results.([]corev2.Mutator)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

// namespaces

func loadNamespacesBatchFn(c client.APIClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for range keys {
			records, err := c.ListNamespaces(&client.ListOptions{})
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadNamespaces(ctx context.Context) ([]corev2.Namespace, error) {
	var records []corev2.Namespace
	loader, err := getLoader(ctx, namespacesLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey("*"))()
	records, ok := results.([]corev2.Namespace)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

// silences

func loadSilencedsBatchFn(c client.APIClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			records, err := c.ListSilenceds(key.String(), "", "", &client.ListOptions{})
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadSilenceds(ctx context.Context, ns string) ([]corev2.Silenced, error) {
	var records []corev2.Silenced
	loader, err := getLoader(ctx, silencedsLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey(ns))()
	records, ok := results.([]corev2.Silenced)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

func contextWithLoaders(ctx context.Context, cfg ServiceConfig, opts ...dataloader.Option) context.Context {
	// Currently all fields are resolved serially, as such we disable batching and
	// rely only on dataloader's cache.
	opts = append([]dataloader.Option{dataloader.WithBatchCapacity(1)}, opts...)
	client := cfg.ClientFactory.NewWithContext(ctx)

	loaders := map[key]*dataloader.Loader{}
	loaders[assetsLoaderKey] = dataloader.NewBatchedLoader(loadAssetsBatchFn(cfg.AssetClient), opts...)
	loaders[checkConfigsLoaderKey] = dataloader.NewBatchedLoader(loadCheckConfigsBatchFn(client), opts...)
	loaders[entitiesLoaderKey] = dataloader.NewBatchedLoader(loadEntitiesBatchFn(client), opts...)
	loaders[eventsLoaderKey] = dataloader.NewBatchedLoader(loadEventsBatchFn(client), opts...)
	loaders[eventFiltersLoaderKey] = dataloader.NewBatchedLoader(loadEventFiltersBatchFn(client), opts...)
	loaders[handlersLoaderKey] = dataloader.NewBatchedLoader(loadHandlersBatchFn(client), opts...)
	loaders[mutatorsLoaderKey] = dataloader.NewBatchedLoader(loadMutatorsBatchFn(client), opts...)
	loaders[namespacesLoaderKey] = dataloader.NewBatchedLoader(loadNamespacesBatchFn(client), opts...)
	loaders[silencedsLoaderKey] = dataloader.NewBatchedLoader(loadSilencedsBatchFn(client), opts...)
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
	if apiErr, ok := err.(client.APIError); ok {
		if apiErr.Code == uint32(actions.PermissionDenied) {
			return nil
		}
	}
	return err
}

// When resolving a field, GraphQL does not consider the absence of a value an
// error; as such we omit the error when the API client returns NotFound or
// Permission denied.
func handleFetchResult(resource interface{}, err error) (interface{}, error) {
	if apiErr, ok := err.(client.APIError); ok {
		if apiErr.Code == uint32(actions.NotFound) || apiErr.Code == uint32(actions.PermissionDenied) {
			return nil, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return resource, err
}
