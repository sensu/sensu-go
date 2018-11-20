package graphql

import (
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"
)

// Assets

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
