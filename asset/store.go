package asset

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// GetAssets retrieves all Assets from the store if contained in the list of asset names
func GetAssets(ctx context.Context, s storev2.Interface, assetList []string) []corev2.Asset {
	assets := make([]corev2.Asset, 0, len(assetList))

	astore := storev2.Of[*corev2.Asset](s)

	for _, assetName := range assetList {
		id := storev2.ID{Namespace: corev2.ContextNamespace(ctx), Name: assetName}
		asset, err := astore.Get(ctx, id)
		if err != nil {
			if _, ok := err.(*store.ErrNotFound); ok {
				logger.WithField("asset", assetName).Info("asset does not exist")
			} else {
				logger.WithField("asset", assetName).WithError(err).Error("error fetching asset from store")
			}
			continue
		}
		assets = append(assets, *asset)
	}

	return assets
}
