// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/coreos/etcd/client"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	testWithEtcdStoreClient(t, func(store store.Store, client client.Client) {
		versionResult, err := store.GetVersion(context.Background(), client)
		assert.NoError(t, err)
		assert.NotEmpty(t, versionResult)
	})
}
