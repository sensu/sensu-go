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
	testWithEtcd(t, func(s store.Store) {
		ext := types.FixtureExtension("frobber")
		ctx := context.WithValue(context.Background(), types.NamespaceKey, ext.Namespace)

		err := s.RegisterExtension(ctx, ext)
		assert.NoError(t, err)

		retrieved, err := s.GetExtension(ctx, "frobber")
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, ext.Name, retrieved.Name)
		assert.Equal(t, ext.URL, retrieved.URL)

		pred := &store.SelectionPredicate{}
		extensions, err := s.GetExtensions(ctx, pred)
		require.NoError(t, err)
		assert.NotEmpty(t, extensions)
		assert.Equal(t, 1, len(extensions))
		assert.Empty(t, pred.Continue)

		// Updating an ext in a nonexistent org should not work
		ext.Namespace = "missing"
		err = s.RegisterExtension(ctx, ext)
		assert.Error(t, err)
	})
}
