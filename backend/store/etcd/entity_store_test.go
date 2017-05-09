package etcd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestEntityStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		entity := &types.Entity{
			ID: "0",
		}
		err := store.UpdateEntity(entity)
		assert.NoError(t, err)
		retrieved, err := store.GetEntityByID(entity.ID)
		assert.NoError(t, err)
		assert.Equal(t, entity.ID, retrieved.ID)
		entities, err := store.GetEntities()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(entities))
		assert.Equal(t, entity.ID, entities[0].ID)
		err = store.DeleteEntity(entity)
		assert.NoError(t, err)
		retrieved, err = store.GetEntityByID(entity.ID)
		assert.Nil(t, retrieved)
		assert.NoError(t, err)
		// Nonexistent entity deletion should return no error.
		err = store.DeleteEntity(entity)
		assert.NoError(t, err)
	})
}
