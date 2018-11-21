package actions

import (
	"crypto/tls"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

// HealthController exposes actions which a viewer can perform
type HealthController struct {
	store               store.HealthStore
	cluster             clientv3.Cluster
	etcdClientTLSConfig *tls.Config
}

// NewHealthController returns new HealthController
func NewHealthController(store store.HealthStore, cluster clientv3.Cluster, etcdClientTLSConfig *tls.Config) HealthController {
	return HealthController{
		store:               store,
		cluster:             cluster,
		etcdClientTLSConfig: etcdClientTLSConfig,
	}
}

// GetClusterHealth returns health information
func (h HealthController) GetClusterHealth(ctx context.Context) *types.HealthResponse {
	return h.store.GetClusterHealth(ctx, h.cluster, h.etcdClientTLSConfig)
}
