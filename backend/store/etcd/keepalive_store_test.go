package etcd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
)

func TestKeepaliveStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		err := store.UpdateKeepalive("default", "entity", 1)
		assert.NoError(t, err)

		retrieved, err := store.GetKeepalive("default", "notfound")
		assert.NoError(t, err)
		assert.Zero(t, retrieved)

		retrieved, err = store.GetKeepalive("default", "entity")
		assert.NoError(t, err)
		assert.NotZero(t, retrieved)
		assert.Equal(t, int64(1), retrieved)
	})
}
