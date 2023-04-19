package mockstore

import (
	"context"
	"crypto/tls"

	v2 "github.com/sensu/core/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// GetClusterHealth ...
func (s *MockStore) GetClusterHealth(ctx context.Context, cluster clientv3.Cluster, etcdClientTLSConfig *tls.Config) *v2.HealthResponse {
	args := s.Called(ctx, cluster, etcdClientTLSConfig)
	return args.Get(0).(*v2.HealthResponse)
}
