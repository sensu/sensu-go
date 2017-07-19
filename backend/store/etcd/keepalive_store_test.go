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
		entity := types.FixtureEntity("entity")

		err := store.UpdateFailingKeepalive(ctx, entity, 1)
		assert.NoError(t, err)

		records, err := store.GetFailingKeepalives(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1, len(records))
	})
}
