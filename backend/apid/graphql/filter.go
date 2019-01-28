package graphql

import "github.com/sensu/sensu-go/types"

// TODO: It would be more ideal to generate the functions in this package

// assets

type assetPredicate func(*types.Asset) bool

func filterAssets(records []types.Asset, filterFn assetPredicate) []*types.Asset {
	relevant := make([]*types.Asset, 0, len(records))

	for i := range records {
		record := records[i]
		if filterFn(&record) {
			relevant = append(relevant, &record)
		}
	}

	return relevant
}

// checks

type checkPredicate func(*types.CheckConfig) bool

func filterChecks(records []types.CheckConfig, filterFn checkPredicate) []*types.CheckConfig {
	relevant := make([]*types.CheckConfig, 0, len(records))

	for i := range records {
		record := records[i]
		if filterFn(&record) {
			relevant = append(relevant, &record)
		}
	}

	return relevant
}

// entities

type entityPredicate func(*types.Entity) bool

func filterEntities(records []types.Entity, filterFn entityPredicate) []*types.Entity {
	relevant := make([]*types.Entity, 0, len(records))

	for i := range records {
		record := records[i]
		if filterFn(&record) {
			relevant = append(relevant, &record)
		}
	}

	return relevant
}

// events

type eventPredicate func(*types.Event) bool

func filterEvents(records []types.Event, filterFn eventPredicate) []*types.Event {
	relevant := make([]*types.Event, 0, len(records))

	for i := range records {
		record := records[i]
		if filterFn(&record) {
			relevant = append(relevant, &record)
		}
	}

	return relevant
}

// handlers

type handlerPredicate func(*types.Handler) bool

func filterHandlers(records []types.Handler, filterFn handlerPredicate) []*types.Handler {
	relevant := make([]*types.Handler, 0, len(records))

	for i := range records {
		record := records[i]
		if filterFn(&record) {
			relevant = append(relevant, &record)
		}
	}

	return relevant
}

// silences

type silencePredicate func(*types.Silenced) bool

func filterSilenceds(records []types.Silenced, filterFn silencePredicate) []*types.Silenced {
	relevant := make([]*types.Silenced, 0, len(records))

	for i := range records {
		record := records[i]
		if filterFn(&record) {
			relevant = append(relevant, &record)
		}
	}

	return relevant
}
