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

func TestEntityStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		entity := types.FixtureEntity("entity")
		ctx := context.WithValue(context.Background(), types.NamespaceKey, entity.Namespace)

		// We should receive an empty slice if no results were found
		entities, err := store.GetEntities(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, entities)

		err = store.UpdateEntity(ctx, entity)
		assert.NoError(t, err)

		retrieved, err := store.GetEntityByName(ctx, entity.Name)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, entity.Name, retrieved.Name)

		entities, err = store.GetEntities(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(entities))
		assert.Equal(t, entity.Name, entities[0].Name)

		err = store.DeleteEntity(ctx, entity)
		assert.NoError(t, err)

		retrieved, err = store.GetEntityByName(ctx, entity.Name)
		assert.Nil(t, retrieved)
		assert.NoError(t, err)

		// Nonexistent entity deletion should return no error.
		err = store.DeleteEntity(ctx, entity)
		assert.NoError(t, err)

		// Updating an enity in a nonexistent org and env should not work
		entity.Namespace = "missing"
		err = store.UpdateEntity(ctx, entity)
		assert.Error(t, err)
	})
}
