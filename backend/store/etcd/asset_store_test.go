package etcd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestAssetStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		asset := types.FixtureAsset("ruby")

		err := store.UpdateAsset(asset)
		assert.NoError(t, err)

		retrieved, err := store.GetAssetByName("default", "ruby")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)

		assert.Equal(t, asset.Name, retrieved.Name)
		assert.Equal(t, asset.URL, retrieved.URL)
		assert.Equal(t, asset.Hash, retrieved.Hash)
		assert.Equal(t, asset.Metadata, retrieved.Metadata)

		assets, err := store.GetAssets("default")
		assert.NoError(t, err)
		assert.NotEmpty(t, assets)
		assert.Equal(t, 1, len(assets))
	})
}
