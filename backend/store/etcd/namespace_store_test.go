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

func TestnamespaceStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		ctx := context.Background()

		// We should receive the default namespace (set in store_test.go)
		namespaces, err := store.ListNamespaces(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(namespaces))

		// We should be able to create a new namespace
		namespace := types.FixtureNamespace("acme")
		err = store.CreateNamespace(ctx, namespace)
		assert.NoError(t, err)

		result, err := store.GetNamespace(ctx, namespace.Name)
		assert.NoError(t, err)
		assert.Equal(t, namespace.Name, result.Name)

		// Missing namespace
		result, err = store.GetNamespace(ctx, "missing")
		assert.NoError(t, err)
		assert.Nil(t, result)

		// Get all namespaces
		namespaces, err = store.ListNamespaces(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, namespaces)
		assert.Equal(t, 2, len(namespaces))

		// Delete a non-empty namespace
		err = store.DeleteNamespace(ctx, namespace.Name)
		assert.Error(t, err)

		// Delete a non-empty namespace w/ roles
		require.NoError(t, store.UpdateRole(ctx, types.FixtureRole("1", namespace.Name)))
		require.NoError(t, store.UpdateRole(ctx, types.FixtureRole("2", "asdf")))
		err = store.DeleteNamespace(ctx, namespace.Name)
		assert.Error(t, err)

		// Delete an empty namespace
		require.NoError(t, store.DeleteRoleByName(ctx, "1"))
		err = store.DeleteNamespace(ctx, namespace.Name)
		assert.NoError(t, err)

		// Delete a missing namespace
		err = store.DeleteNamespace(ctx, "missing")
		assert.Error(t, err)

		// Get again all namespaces
		namespaces, err = store.ListNamespaces(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(namespaces))
	})
}
