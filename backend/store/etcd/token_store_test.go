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
		err = store.DeleteTokens(claims.Subject, []string{claims.Id})
		assert.NoError(t, err)
		_, err = store.GetToken(claims.Subject, claims.Id)
		assert.Error(t, err)
	})
}
