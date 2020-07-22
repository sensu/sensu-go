// +build integration,!race

package etcd

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"go.etcd.io/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
)

func TestGetClusterHealth(t *testing.T) {
	testWithEtcdClient(t, func(s store.Store, client *clientv3.Client) {
		healthResult := s.GetClusterHealth(context.Background(), client.Cluster, (*tls.Config)(nil))
		assert.Empty(t, healthResult.ClusterHealth[0].Err)
	})
}

func TestGetClusterHealthTimeout(t *testing.T) {
	testWithEtcdClient(t, func(s store.Store, client *clientv3.Client) {
		result := s.GetClusterHealth(context.WithValue(context.Background(), store.ContextKeyTimeout, time.Nanosecond), client.Cluster, (*tls.Config)(nil))
		assert.NotEmpty(t, result.ClusterHealth[0].Err)
	})
}
