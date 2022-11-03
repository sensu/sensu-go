package actions

import (
	"crypto/tls"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"go.etcd.io/etcd/client/v3"
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
func (h HealthController) GetClusterHealth(ctx context.Context) *corev2.HealthResponse {
	return h.store.GetClusterHealth(ctx, h.cluster, h.etcdClientTLSConfig)
}
