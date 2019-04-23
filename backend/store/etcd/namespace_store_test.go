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
	testWithEtcd(t, func(s store.Store) {
		ctx := context.Background()

		// We should receive the default namespace (set in store_test.go)
		pred := &store.SelectionPredicate{}
		namespaces, err := s.ListNamespaces(ctx, pred)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(namespaces))

		// We should be able to create a new namespace
		namespace := types.FixtureNamespace("acme")
		err = s.CreateNamespace(ctx, namespace)
		assert.NoError(t, err)

		result, err := s.GetNamespace(ctx, namespace.Name)
		assert.NoError(t, err)
		assert.Equal(t, namespace.Name, result.Name)

		// Missing namespace
		result, err = s.GetNamespace(ctx, "missing")
		assert.NoError(t, err)
		assert.Nil(t, result)

		// Get all namespaces
		namespaces, err = s.ListNamespaces(ctx, pred)
		assert.NoError(t, err)
		assert.NotEmpty(t, namespaces)
		assert.Equal(t, 2, len(namespaces))

		// Delete a non-empty namespace
		err = s.DeleteNamespace(ctx, namespace.Name)
		assert.Error(t, err)

		// Delete a non-empty namespace w/ roles
		require.NoError(t, s.UpdateRole(ctx, types.FixtureRole("1", namespace.Name)))
		require.NoError(t, s.UpdateRole(ctx, types.FixtureRole("2", "asdf")))
		err = s.DeleteNamespace(ctx, namespace.Name)
		assert.Error(t, err)

		// Delete an empty namespace
		require.NoError(t, s.DeleteRole(ctx, "1"))
		err = s.DeleteNamespace(ctx, namespace.Name)
		assert.NoError(t, err)

		// Delete a missing namespace
		err = s.DeleteNamespace(ctx, "missing")
		assert.Error(t, err)

		// Get again all namespaces
		namespaces, err = s.ListNamespaces(ctx, pred)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(namespaces))
	})
}
