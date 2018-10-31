// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtensionStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		ext := types.FixtureExtension("frobber")
		ctx := context.WithValue(context.Background(), types.NamespaceKey, ext.Namespace)

		err := store.RegisterExtension(ctx, ext)
		assert.NoError(t, err)

		retrieved, err := store.GetExtension(ctx, "frobber")
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, ext.Name, retrieved.Name)
		assert.Equal(t, ext.URL, retrieved.URL)

		extensions, err := store.GetExtensions(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, extensions)
		assert.Equal(t, 1, len(extensions))

		// Updating an ext in a nonexistent org should not work
		ext.Namespace = "missing"
		err = store.RegisterExtension(ctx, ext)
		assert.Error(t, err)
	})
}
