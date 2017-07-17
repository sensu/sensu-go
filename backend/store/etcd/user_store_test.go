package etcd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestUserStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		password := "P@ssw0rd!"

		// We should receive an empty array if no users exist
		users, err := store.GetUsers()
		assert.NoError(t, err)
		assert.Empty(t, users)

		user := types.FixtureUser("foo")
		user.Password = password
		err = store.CreateUser(user)
		assert.NoError(t, err)

		// The password should be hashed
		result, err := store.GetUser("foo")
		assert.NoError(t, err)
		assert.NotEqual(t, password, result.Password)

		// Successful authentication
		_, err = store.AuthenticateUser("foo", password)
		assert.NoError(t, err)

		// Unsuccessful authentication with wrong password
		_, err = store.AuthenticateUser("foo", "foo")
		assert.Error(t, err)

		// User already exist
		err = store.CreateUser(user)
		assert.Error(t, err)

		mockedUser := types.FixtureUser("bar")
		mockedUser.Password = password
		err = store.UpdateUser(mockedUser)
		assert.NoError(t, err)

		result, err = store.GetUser(mockedUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, mockedUser.Username, result.Username)

		// Missing user
		_, err = store.GetUser("missingUser")
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

		// Authentication should be unsuccessful with a disabled user
		_, err = store.AuthenticateUser("bar", password)
		assert.Error(t, err)

		// Disable a missing user
		err = store.DeleteUserByName("missingUser")
		assert.Error(t, err)

		// The deleted (disabled) user should not be returned
		// Get all users
		users, err = store.GetUsers()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(users))
	})
}

func TestCheckPassword(t *testing.T) {
	hash := "$2a$10$iyYyGmveS9dcYp5DHMbOm.LShX806vB0ClzoPyt1TIgkZ9KQ62cOO"
	password := "P@ssw0rd!"

	assert.False(t, checkPassword(hash, "foo"))
	assert.True(t, checkPassword(hash, password))
}

func TestHashPassword(t *testing.T) {
	password := "P@ssw0rd!"

	hash, err := hashPassword(password)
	assert.NotEqual(t, password, hash)
	assert.NoError(t, err)
}
