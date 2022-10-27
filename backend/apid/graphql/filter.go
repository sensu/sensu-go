package graphql

import corev2 "github.com/sensu/core/v2"

// TODO: It would be more ideal to generate the functions in this package

// assets

type assetPredicate func(*corev2.Asset) bool

func filterAssets(records []*corev2.Asset, filterFn assetPredicate) []*corev2.Asset {
	relevant := make([]*corev2.Asset, 0, len(records))

	for i := range records {
		record := records[i]
		if filterFn(record) {
			relevant = append(relevant, record)
		}
	}

	return relevant
}

// entities

type entityPredicate func(*corev2.Entity) bool

func filterEntities(records []*corev2.Entity, filterFn entityPredicate) []*corev2.Entity {
	relevant := make([]*corev2.Entity, 0, len(records))

	for _, record := range records {
		if filterFn(record) {
			relevant = append(relevant, record)
		}
	}

	return relevant
}

// events

type eventPredicate func(*corev2.Event) bool

func filterEvents(records []*corev2.Event, filterFn eventPredicate) []*corev2.Event {
	relevant := make([]*corev2.Event, 0, len(records))

	for _, record := range records {
		if filterFn(record) {
			relevant = append(relevant, record)
		}
	}

	return relevant
}

// handlers

type handlerPredicate func(*corev2.Handler) bool

func filterHandlers(records []*corev2.Handler, filterFn handlerPredicate) []*corev2.Handler {
	relevant := make([]*corev2.Handler, 0, len(records))

	for _, record := range records {
		if filterFn(record) {
			relevant = append(relevant, record)
		}
	}

	return relevant
}

// silences

type silencePredicate func(*corev2.Silenced) bool

func filterSilenceds(records []*corev2.Silenced, filterFn silencePredicate) []*corev2.Silenced {
	relevant := make([]*corev2.Silenced, 0, len(records))

	for _, record := range records {
		if filterFn(record) {
			relevant = append(relevant, record)
		}
	}

	return relevant
}
