// +build integration,!race

package etcd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	utilbytes "github.com/sensu/sensu-go/util/bytes"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticationStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		// Secret does not exist
		_, err := store.GetJWTSecret()
		assert.Error(t, err)

		// Create a secret
		secret, _ := utilbytes.Random(32)
		err = store.CreateJWTSecret(secret)
		assert.NoError(t, err)

		// Retrieve the secret
		result, err := store.GetJWTSecret()
		assert.NoError(t, err)
		assert.Equal(t, secret, result)

		// We should not be able to create it again
		err = store.CreateJWTSecret(secret)
		assert.Error(t, err)

		// We should be able to update it
		newSecret, _ := utilbytes.Random(32)
		err = store.UpdateJWTSecret(newSecret)
		assert.NoError(t, err)

		// The old and new secrets should not match
		result, err = store.GetJWTSecret()
		assert.NoError(t, err)
		assert.NotEqual(t, result, secret)
	})
}
