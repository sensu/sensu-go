// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestAssetStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		asset := types.FixtureAsset("ruby")
		ctx := context.WithValue(context.Background(), types.NamespaceKey, asset.Namespace)

		err := store.UpdateAsset(ctx, asset)
		assert.NoError(t, err)

		retrieved, err := store.GetAssetByName(ctx, "ruby")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)

		assert.Equal(t, asset.Name, retrieved.Name)
		assert.Equal(t, asset.URL, retrieved.URL)
		assert.Equal(t, asset.Sha512, retrieved.Sha512)
		assert.Equal(t, asset.Metadata, retrieved.Metadata)

		assets, err := store.GetAssets(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, assets)
		assert.Equal(t, 1, len(assets))

		// Updating an asset in a nonexistent org should not work
		asset.Namespace = "missing"
		err = store.UpdateAsset(ctx, asset)
		assert.Error(t, err)
	})
}
