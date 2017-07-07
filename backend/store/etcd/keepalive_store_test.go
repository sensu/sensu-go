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
		ctx := context.WithValue(context.Background(), types.OrganizationKey, "default")

		err := store.UpdateKeepalive(ctx, "entity", 1)
		assert.NoError(t, err)

		retrieved, err := store.GetKeepalive(ctx, "notfound")
		assert.NoError(t, err)
		assert.Zero(t, retrieved)

		retrieved, err = store.GetKeepalive(ctx, "entity")
		assert.NoError(t, err)
		assert.NotZero(t, retrieved)
		assert.Equal(t, int64(1), retrieved)
	})
}
