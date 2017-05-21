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
		user.Password = "P@ssw0rd!"
		err := store.CreateUser(user)
		assert.NoError(t, err)

		// User already exist
		err = store.CreateUser(user)
		assert.Error(t, err)

		mockedUser := types.FixtureUser("bar")
		err = store.UpdateUser(mockedUser)
		assert.NoError(t, err)

		result, err := store.GetUser(mockedUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, mockedUser.Username, result.Username)

		// Missing user
		_, err = store.GetUser("foobar")
		assert.Error(t, err)

		// Get all users
		users, err := store.GetUsers()
		assert.NoError(t, err)
		assert.NotEmpty(t, users)
		assert.Equal(t, 2, len(users))
	})
}
