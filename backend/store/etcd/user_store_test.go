package etcd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestUserStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		user := types.FixtureUser("foo")
		err := store.UpdateUser(user)
		assert.NoError(t, err)
	})
}
