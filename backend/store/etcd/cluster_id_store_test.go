// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
)

func TestClusterIDStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		ctx := context.Background()

		// We should receive an empty string if none exists
		id, err := store.GetClusterID(ctx)
		assert.Error(t, err)
		assert.Empty(t, id)

		// We should not receive an error adding a uuid
		u := uuid.New().String()
		err = store.CreateClusterID(ctx, u)
		assert.NoError(t, err)

		// We should get a matching uuid
		id, err = store.GetClusterID(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, id)
		assert.Equal(t, id, u)

		// We should not receive an error updating the uuid
		u = uuid.New().String()
		err = store.CreateClusterID(ctx, u)
		assert.NoError(t, err)

		// We should get a matching uuid
		id, err = store.GetClusterID(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, id)
		assert.Equal(t, id, u)
	})
}
