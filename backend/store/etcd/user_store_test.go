// +build integration,!race

package etcd

import (
	"context"
	"fmt"
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
	testWithEtcd(t, func(store store.Store) {
		password := "P@ssw0rd!"
		passwordDigest, err := bcrypt.HashPassword(password)
		require.NoError(t, err)

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
		user.Password = passwordDigest
		err = store.CreateUser(user)
		assert.NoError(t, err)

		// The user should be fetchable
		result, err := store.GetUser(ctx, "foo")
		assert.NoError(t, err)
		assert.NotNil(t, result)

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
		mockedUser.Password = passwordDigest
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
		claims := corev2.FixtureClaims("bar", nil)
		token, _, _ := jwt.AccessToken(claims)
		err = store.AllowTokens(token)
		assert.NoError(t, err)

		// Disable a user that does not exist
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
		users, _, err = store.GetAllUsers(0, "")
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))
	})
}

func TestGetAllUsersPagination(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		for i := 1; i <= 21; i++ {
			// We force the object name to be 2 digits "wide" in order to
			// have a "natural" lexicographic order: 01, 02, ... instead of 1,
			// 11, ...
			objectName := fmt.Sprintf("%.2d", i)
			object := corev2.FixtureUser(objectName)

			if err := store.CreateUser(object); err != nil {
				t.Fatal(err)
			}
		}

		ctx := context.Background()
		t.Run("paginate through users", func(t *testing.T) {
			testGetAllUsersPagination(t, ctx, store, 10, 21)
		})

		t.Run("page size equals one", func(t *testing.T) {
			testGetAllUsersPagination(t, ctx, store, 1, 21)
		})

		t.Run("page size bigger than set size", func(t *testing.T) {
			testGetAllUsersPagination(t, ctx, store, 1337, 21)
		})
	})
}

func testGetAllUsersPagination(t *testing.T, ctx context.Context, etcd store.Store, pageSize, setSize int) {
	nFullPages := setSize / pageSize
	nLeftovers := setSize % pageSize

	continueToken := ""
	for i := 0; i < nFullPages; i++ {
		objects, nextContinueToken, err := etcd.GetAllUsers(int64(pageSize), continueToken)
		if err != nil {
			t.Fatal(err)
		}

		if len(objects) != pageSize {
			t.Fatalf("Expected page %d to have %d objects but got %d", i, pageSize, len(objects))
		}

		offset := i * pageSize
		for j, object := range objects {
			n := ((offset + j) % setSize) + 1
			expected := fmt.Sprintf("%.2d", n)

			if object.Username != expected {
				t.Fatalf("Expected %s, got %s", expected, object.Username)
			}
		}

		continueToken = nextContinueToken
	}

	// Check the last page, supposed to hold nLeftovers objects
	if nLeftovers > 0 {
		objects, nextContinueToken, err := etcd.GetAllUsers(int64(pageSize), continueToken)
		if err != nil {
			t.Fatal(err)
		}

		if len(objects) != nLeftovers {
			t.Fatalf("Expected last page with %d objects, got %d", nLeftovers, len(objects))
		}

		if nextContinueToken != "" {
			t.Fatalf("Expected next continue token to be \"\", got %s", nextContinueToken)
		}

		offset := pageSize * nFullPages
		for j, object := range objects {
			n := ((offset + j) % setSize) + 1
			expected := fmt.Sprintf("%.2d", n)

			if object.Username != expected {
				t.Fatalf("Expected %s, got %s", expected, object.Username)
			}
		}
	}
}
