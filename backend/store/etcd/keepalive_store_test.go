package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestKeepaliveStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		entity := types.FixtureEntity("entity")
		ctx := context.WithValue(context.Background(), types.NamespaceKey, entity.Namespace)

		err := store.UpdateFailingKeepalive(ctx, entity, 1)
		assert.NoError(t, err)

		records, err := store.GetFailingKeepalives(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1, len(records))

		// Updating a keepalive in a nonexistent org and env should not work
		entity.Namespace = "missing"
		err = store.UpdateFailingKeepalive(ctx, entity, 1)
		assert.Error(t, err)
	})
}
