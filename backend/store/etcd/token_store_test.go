package etcd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
)

func TestTokensStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		// Generate a dummy access token
		token, _, _ := jwt.AccessToken("foo")
		claims, _ := jwt.GetClaims(token)

		// Store the access token
		err := store.CreateToken(claims)
		assert.NoError(t, err)

		// Retrieve the stored token
		result, err := store.GetToken(claims.Subject, claims.Id)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Delete the stored token
		err = store.DeleteToken(claims.Subject, claims.Id)
		assert.NoError(t, err)
		_, err = store.GetToken(claims.Subject, claims.Id)
		assert.Error(t, err)

		// Generate multiple tokens for bar user
		token1, _, _ := jwt.AccessToken("bar")
		claims1, _ := jwt.GetClaims(token1)
		err = store.CreateToken(claims1)
		assert.NoError(t, err)
		token2, _, _ := jwt.AccessToken("bar")
		claims2, _ := jwt.GetClaims(token2)
		err = store.CreateToken(claims2)
		assert.NoError(t, err)

		// Make sure the tokens exists
		result, err = store.GetToken(claims1.Subject, claims1.Id)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		result, err = store.GetToken(claims2.Subject, claims2.Id)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Delete all tokens for user bar
		err = store.DeleteTokensByUsername("bar")
		assert.NoError(t, err)

		// Make sure the tokens do not exist anymore
		_, err = store.GetToken(claims1.Subject, claims1.Id)
		assert.Error(t, err)
		_, err = store.GetToken(claims2.Subject, claims2.Id)
		assert.Error(t, err)

	})
}
