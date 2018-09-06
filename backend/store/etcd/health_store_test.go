// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
)

func TestGetClusterHealth(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		healthResult := store.GetClusterHealth(context.Background())
		assert.NoError(t, healthResult.ClusterHealth[0].Err)
	})
}
