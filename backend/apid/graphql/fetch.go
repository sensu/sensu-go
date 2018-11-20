package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"
)

// assets

type assetPredicate func(*types.Asset) bool

func fetchAssets(c client.APIClient, ns string, filter assetPredicate) ([]*types.Asset, error) {
	records, err := c.ListAssets(ns)
	relevant := make([]*types.Asset, 0, len(records))
	if err != nil {
		return relevant, err
	}

	if filter == nil {
		filter = func(*types.Asset) bool { return true }
	}

	for _, record := range records {
		if filter(&record) {
			relevant = append(relevant, &record)
		}
	}

	return relevant, nil
}

// entitys

type entityPredicate func(*types.Entity) bool

func fetchEntities(c client.APIClient, ns string, filter entityPredicate) ([]*types.Entity, error) {
	records, err := c.ListEntities(ns)
	relevant := make([]*types.Entity, 0, len(records))
	if err != nil {
		return relevant, err
	}

	if filter == nil {
		filter = func(*types.Entity) bool { return true }
	}

	for _, record := range records {
		if filter(&record) {
			relevant = append(relevant, &record)
		}
	}

	return relevant, nil
}

// events

type eventPredicate func(*types.Event) bool

func fetchEvents(c client.APIClient, ns string, filter eventPredicate) ([]*types.Event, error) {
	records, err := c.ListEvents(ns)
	relevant := make([]*types.Event, 0, len(records))
	if err != nil {
		return relevant, err
	}

	if filter == nil {
		filter = func(*types.Event) bool { return true }
	}

	for _, record := range records {
		if filter(&record) {
			relevant = append(relevant, &record)
		}
	}

	return relevant, nil
}

// handlers

type handlerPredicate func(*types.Handler) bool

func fetchHandlers(c client.APIClient, ns string, filter handlerPredicate) ([]*types.Handler, error) {
	records, err := c.ListHandlers(ns)
	relevant := make([]*types.Handler, 0, len(records))
	if err != nil {
		return relevant, err
	}

	if filter == nil {
		filter = func(*types.Handler) bool { return true }
	}

	for _, record := range records {
		if filter(&record) {
			relevant = append(relevant, &record)
		}
	}

	return relevant, nil
}

// silences

type silencePredicate func(*types.Silenced) bool

func fetchSilenceds(c client.APIClient, ns string, filter silencePredicate) ([]*types.Silenced, error) {
	records, err := c.ListSilenceds(ns, "", "")
	relevant := make([]*types.Silenced, 0, len(records))
	if err != nil {
		return relevant, err
	}

	if filter == nil {
		filter = func(*types.Silenced) bool { return true }
	}

	for _, record := range records {
		if filter(&record) {
			relevant = append(relevant, &record)
		}
	}

	return relevant, nil
}

// When resolving a field, GraphQL does not consider the absence of a value an
// error; as such we omit the error when the API client returns NotFound.
func wrapFetch(resource interface{}, err error) (interface{}, error) {
	if apiErr, ok := err.(client.APIError); ok {
		if apiErr.Code == uint32(actions.NotFound) { // TODO: Reference error codes
			return nil, nil
		}
	}
	return resource, err
}
