package asset

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"go.uber.org/multierr"
)

// GetAssets retrieves all Assets from the store if contained in the list of asset names
func GetAssets(ctx context.Context, store store.Store, assetList []string) ([]types.Asset, error) {
	assets := []types.Asset{}

	var errors error
	for _, assetName := range assetList {
		asset, err := store.GetAssetByName(ctx, assetName)
		if err != nil {
			logger.WithField("asset", assetName).WithError(err).Error("error fetching asset from store")
			errors = multierr.Append(errors, err)
		} else if asset == nil {
			logger.WithField("asset", assetName).Info("asset does not exist")
		} else {
			assets = append(assets, *asset)
		}
	}

	return assets, errors
}
