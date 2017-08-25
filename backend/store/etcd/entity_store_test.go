package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestEntityStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		entity := types.FixtureEntity("entity")
		ctx := context.WithValue(context.Background(), types.OrganizationKey, entity.Organization)
		ctx = context.WithValue(ctx, types.EnvironmentKey, entity.Environment)

		// We should receive an empty slice if no results were found
		entities, err := store.GetEntities(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, entities)

		err = store.UpdateEntity(ctx, entity)
		assert.NoError(t, err)

		retrieved, err := store.GetEntityByID(ctx, entity.ID)
		assert.NoError(t, err)
		assert.Equal(t, entity.ID, retrieved.ID)

		entities, err = store.GetEntities(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(entities))
		assert.Equal(t, entity.ID, entities[0].ID)

		err = store.DeleteEntity(ctx, entity)
		assert.NoError(t, err)

		retrieved, err = store.GetEntityByID(ctx, entity.ID)
		assert.Nil(t, retrieved)
		assert.NoError(t, err)

		// Nonexistent entity deletion should return no error.
		err = store.DeleteEntity(ctx, entity)
		assert.NoError(t, err)

		// Updating an enity in an inexistant org and env should not work
		entity.Organization = "missing"
		entity.Environment = "missing"
		err = store.UpdateEntity(ctx, entity)
		assert.Error(t, err)
	})
}
