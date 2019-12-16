// +build integration,!race

package etcd

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
)

func TestGetClusterHealth(t *testing.T) {
	testWithEtcdClient(t, func(store store.Store, client *clientv3.Client) {
		healthResult := store.GetClusterHealth(context.Background(), client.Cluster, (*tls.Config)(nil))
		assert.Empty(t, healthResult.ClusterHealth[0].Err)
	})
}

func TestGetClusterHealthTimeout(t *testing.T) {
	testWithEtcdClient(t, func(store store.Store, client *clientv3.Client) {
		result := store.GetClusterHealth(context.WithValue(context.Background(), "timeout", time.Nanosecond), client.Cluster, (*tls.Config)(nil))
		assert.NotEmpty(t, result.ClusterHealth[0].Err)
	})
}
