// +build integration,!race

package etcd

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestUserStorage(t *testing.T) {
	testWithEtcd(t, func(s store.Store) {
		password := "P@ssw0rd!"
		passwordDigest, err := bcrypt.HashPassword(password)
		require.NoError(t, err)

		ctx, cancel := context.WithDeadline(
			context.Background(),
			time.Now().Add(20*time.Second),
		)
		defer cancel()

		// We should receive an empty array if no users exist
		users, err := s.GetUsers()
		assert.NoError(t, err)
		assert.Empty(t, users)

		user := types.FixtureUser("foo")
		user.Password = passwordDigest
		err = s.CreateUser(user)
		assert.NoError(t, err)

		// The user should be fetchable
		result, err := s.GetUser(ctx, "foo")
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Successful authentication
		_, err = s.AuthenticateUser(ctx, "foo", password)
		assert.NoError(t, err)

		// Unsuccessful authentication with wrong password
		_, err = s.AuthenticateUser(ctx, "foo", "foo")
		assert.Error(t, err)

		// User already exist
		err = s.CreateUser(user)
		assert.Error(t, err)

		mockedUser := types.FixtureUser("bar")
		mockedUser.Password = passwordDigest
		err = s.UpdateUser(mockedUser)
		assert.NoError(t, err)

		result, err = s.GetUser(ctx, mockedUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, mockedUser.Username, result.Username)

		// Missing user
		missingUser, err := s.GetUser(ctx, "missingUser")
		assert.NoError(t, err)
		assert.Nil(t, missingUser)

		// Get all users
		users, err = s.GetUsers()
		assert.NoError(t, err)
		assert.NotEmpty(t, users)
		assert.Equal(t, 2, len(users))

		// Generate a token for the bar user
		claims := corev2.FixtureClaims("bar", nil)
		token, _, _ := jwt.AccessToken(claims)
		err = s.AllowTokens(token)
		assert.NoError(t, err)

		// Disable a user that does not exist
		err = s.DeleteUser(ctx, &types.User{Username: "Frankieie"})
		assert.NoError(t, err)

		// Ensure that a user with that name wasn't created
		baduser, err := s.GetUser(ctx, "Frankieie")
		assert.NoError(t, err)
		assert.Nil(t, baduser)

		// Disable a user, which also removes all issued tokens
		err = s.DeleteUser(ctx, mockedUser)
		assert.NoError(t, err)

		// Make sure the user is now disabled
		disabledUser, _ := s.GetUser(ctx, mockedUser.Username)
		assert.True(t, disabledUser.Disabled)

		// Make sure the token was revoked
		_, err = s.GetToken(claims.Subject, claims.Id)
		assert.Error(t, err)

		// Authentication should be unsuccessful with a disabled user
		_, err = s.AuthenticateUser(ctx, mockedUser.Username, password)
		assert.Error(t, err)

		// The deleted (disabled) user should not be returned
		// Get all users
		users, err = s.GetUsers()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(users))

		// Disabled user should appear when fetching all users
		pred := &store.SelectionPredicate{}
		users, err = s.GetAllUsers(pred)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))
	})
}
