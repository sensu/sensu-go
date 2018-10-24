// +build integration,!race

package etcd

import (
	"context"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestUserStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		password := "P@ssw0rd!"
		ctx, cancel := context.WithDeadline(
			context.Background(),
			time.Now().Add(20*time.Second),
		)
		defer cancel()

		// We should receive an empty array if no users exist
		users, err := store.GetUsers()
		assert.NoError(t, err)
		assert.Empty(t, users)

		user := types.FixtureUser("foo")
		user.Password = password
		err = store.CreateUser(user)
		assert.NoError(t, err)

		// The password should be hashed
		result, err := store.GetUser(ctx, "foo")
		assert.NoError(t, err)
		assert.NotEqual(t, password, result.Password)

		// Successful authentication
		_, err = store.AuthenticateUser(ctx, "foo", password)
		assert.NoError(t, err)

		// Unsuccessful authentication with wrong password
		_, err = store.AuthenticateUser(ctx, "foo", "foo")
		assert.Error(t, err)

		// User already exist
		err = store.CreateUser(user)
		assert.Error(t, err)

		mockedUser := types.FixtureUser("bar")
		mockedUser.Password = password
		err = store.UpdateUser(mockedUser)
		assert.NoError(t, err)

		result, err = store.GetUser(ctx, mockedUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, mockedUser.Username, result.Username)

		// Missing user
		missingUser, err := store.GetUser(ctx, "missingUser")
		assert.NoError(t, err)
		assert.Nil(t, missingUser)

		// Get all users
		users, err = store.GetUsers()
		assert.NoError(t, err)
		assert.NotEmpty(t, users)
		assert.Equal(t, 2, len(users))

		// Generate a token for the bar user
		userBar := &types.User{Username: "bar"}
		token, _, _ := jwt.AccessToken(userBar)
		claims, _ := jwt.GetClaims(token)
		err = store.CreateToken(claims)
		assert.NoError(t, err)

		// Disable a user a user that does not exist
		err = store.DeleteUser(ctx, &types.User{Username: "Frankieie"})
		assert.NoError(t, err)

		// Ensure that a user with that name wasn't created
		baduser, err := store.GetUser(ctx, "Frankieie")
		assert.NoError(t, err)
		assert.Nil(t, baduser)

		// Disable a user, which also removes all issued tokens
		err = store.DeleteUser(ctx, mockedUser)
		assert.NoError(t, err)

		// Make sure the user is now disabled
		disabledUser, _ := store.GetUser(ctx, mockedUser.Username)
		assert.True(t, disabledUser.Disabled)

		// Make sure the token was revoked
		_, err = store.GetToken(claims.Subject, claims.Id)
		assert.Error(t, err)

		// Authentication should be unsuccessful with a disabled user
		_, err = store.AuthenticateUser(ctx, mockedUser.Username, password)
		assert.Error(t, err)

		// The deleted (disabled) user should not be returned
		// Get all users
		users, err = store.GetUsers()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(users))

		// Disabled user should appear when fetching all users
		users, err = store.GetAllUsers()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))
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
