package mockstore

import (
	"context"
	"crypto/tls"

	"go.etcd.io/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

// GetClusterHealth ...
func (s *MockStore) GetClusterHealth(ctx context.Context, cluster clientv3.Cluster, etcdClientTLSConfig *tls.Config) *types.HealthResponse {
	args := s.Called(ctx, cluster, etcdClientTLSConfig)
	return args.Get(0).(*types.HealthResponse)
}
