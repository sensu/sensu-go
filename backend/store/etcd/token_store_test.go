// +build integration,!race

package etcd

import (
	"testing"

	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
)

func TestTokensStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		// Generate dummy claims
		claims := v2.FixtureClaims("foo", nil)
		token, _, _ := jwt.AccessToken(claims)

		// Add the token
		err := store.AllowTokens(token)
		assert.NoError(t, err)

		// Retrieve the stored token
		result, err := store.GetToken(claims.Subject, claims.Id)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Revoke the token
		err = store.RevokeTokens(claims)
		assert.NoError(t, err)
		_, err = store.GetToken(claims.Subject, claims.Id)
		assert.Error(t, err)
	})
}
