package etcd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestUserStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		// We should receive an empty array if no users exist
		users, err := store.GetUsers()
		assert.NoError(t, err)
		assert.Empty(t, users)

		user := types.FixtureUser("foo")
		user.Password = "P@ssw0rd!"
		err = store.CreateUser(user)
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
		users, err = store.GetUsers()
		assert.NoError(t, err)
		assert.NotEmpty(t, users)
		assert.Equal(t, 2, len(users))

		// Disable a user
		err = store.DeleteUserByName("bar")
		assert.NoError(t, err)

		result, err = store.GetUser("bar")
		assert.NoError(t, err)
		assert.True(t, result.Disabled)

		// Disable a missing user
		err = store.DeleteUserByName("foobar")
		assert.Error(t, err)

		// The deleted (disabled) user should not be returned
		// Get all users
		users, err = store.GetUsers()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(users))
	})
}
