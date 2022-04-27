package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestAssetStorage(t *testing.T) {
	testWithEtcd(t, func(s store.Store) {
		asset := types.FixtureAsset("ruby")
		ctx := context.WithValue(context.Background(), types.NamespaceKey, asset.Namespace)

		err := s.UpdateAsset(ctx, asset)
		assert.NoError(t, err)

		retrieved, err := s.GetAssetByName(ctx, "ruby")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)

		assert.Equal(t, asset.Name, retrieved.Name)
		assert.Equal(t, asset.URL, retrieved.URL)
		assert.Equal(t, asset.Sha512, retrieved.Sha512)

		pred := &store.SelectionPredicate{}
		assets, err := s.GetAssets(ctx, pred)
		assert.NoError(t, err)
		assert.NotEmpty(t, assets)
		assert.Equal(t, 1, len(assets))
		assert.Empty(t, pred.Continue)

		// Updating an asset in a nonexistent org should not work
		asset.Namespace = "missing"
		err = s.UpdateAsset(ctx, asset)
		assert.Error(t, err)
	})
}
