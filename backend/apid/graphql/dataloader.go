package graphql

import (
	"context"
	"errors"

	"github.com/graph-gophers/dataloader"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"
)

type key int

const (
	loadersKey key = iota
	assetsLoaderKey
	checkConfigsLoaderKey
	entitiesLoaderKey
	eventsLoaderKey
	handlersLoaderKey
	namespacesLoaderKey
	silencedsLoaderKey
)

var (
	errLoadersNotFound        = errors.New("loaders was not found inside context")
	errLoaderNotFound         = errors.New("loader was not found")
	errUnexpectedLoaderResult = errors.New("loader returned unexpected result")
)

// assets

func loadAssetsBatchFn(client client.APIClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			records, err := client.ListAssets(key.String())
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadAssets(ctx context.Context, ns string) ([]types.Asset, error) {
	var records []types.Asset
	loader, err := getLoader(ctx, assetsLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey(ns))()
	records, ok := results.([]types.Asset)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

// checks

func loadCheckConfigsBatchFn(client client.APIClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			records, err := client.ListChecks(key.String())
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadCheckConfigs(ctx context.Context, ns string) ([]types.CheckConfig, error) {
	var records []types.CheckConfig
	loader, err := getLoader(ctx, checkConfigsLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey(ns))()
	records, ok := results.([]types.CheckConfig)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

// entities

func loadEntitiesBatchFn(client client.APIClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			records, err := client.ListEntities(key.String())
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadEntities(ctx context.Context, ns string) ([]types.Entity, error) {
	var records []types.Entity
	loader, err := getLoader(ctx, entitiesLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey(ns))()
	records, ok := results.([]types.Entity)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

// events

func loadEventsBatchFn(client client.APIClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			records, err := client.ListEvents(key.String())
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadEvents(ctx context.Context, ns string) ([]types.Event, error) {
	var records []types.Event
	loader, err := getLoader(ctx, eventsLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey(ns))()
	records, ok := results.([]types.Event)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

// handlers

func loadHandlersBatchFn(client client.APIClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			records, err := client.ListHandlers(key.String())
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadHandlers(ctx context.Context, ns string) ([]types.Handler, error) {
	var records []types.Handler
	loader, err := getLoader(ctx, handlersLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey(ns))()
	records, ok := results.([]types.Handler)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

// namespaces

func loadNamespacesBatchFn(client client.APIClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for range keys {
			records, err := client.ListNamespaces()
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadNamespaces(ctx context.Context) ([]types.Namespace, error) {
	var records []types.Namespace
	loader, err := getLoader(ctx, namespacesLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey("*"))()
	records, ok := results.([]types.Namespace)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

// silences

func loadSilencedsBatchFn(client client.APIClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		results := make([]*dataloader.Result, 0, len(keys))
		for _, key := range keys {
			records, err := client.ListSilenceds(key.String(), "", "")
			result := &dataloader.Result{Data: records, Error: handleListErr(err)}
			results = append(results, result)
		}
		return results
	}
}

func loadSilenceds(ctx context.Context, ns string) ([]types.Silenced, error) {
	var records []types.Silenced
	loader, err := getLoader(ctx, silencedsLoaderKey)
	if err != nil {
		return records, err
	}

	results, err := loader.Load(ctx, dataloader.StringKey(ns))()
	records, ok := results.([]types.Silenced)
	if err == nil && !ok {
		err = errUnexpectedLoaderResult
	}
	return records, err
}

func contextWithLoaders(ctx context.Context, client client.APIClient, opts ...dataloader.Option) context.Context {
	// Currently all fields are resolved serially, as such we disable batching and
	// rely only on dataloader's cache.
	opts = append([]dataloader.Option{dataloader.WithBatchCapacity(1)}, opts...)

	loaders := map[key]*dataloader.Loader{}
	loaders[assetsLoaderKey] = dataloader.NewBatchedLoader(loadAssetsBatchFn(client), opts...)
	loaders[checkConfigsLoaderKey] = dataloader.NewBatchedLoader(loadCheckConfigsBatchFn(client), opts...)
	loaders[entitiesLoaderKey] = dataloader.NewBatchedLoader(loadEntitiesBatchFn(client), opts...)
	loaders[eventsLoaderKey] = dataloader.NewBatchedLoader(loadEventsBatchFn(client), opts...)
	loaders[handlersLoaderKey] = dataloader.NewBatchedLoader(loadHandlersBatchFn(client), opts...)
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
