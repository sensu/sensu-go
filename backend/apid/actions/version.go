package actions

import (
	etcdVersion "github.com/coreos/etcd/version"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/version"
	"golang.org/x/net/context"
)

// VersionController exposes actions which a viewer can perform
type VersionController struct {
	clusterVersion string
}

// NewVersionController returns a new VersionController
func NewVersionController(clusterVersion string) VersionController {
	return VersionController{
		clusterVersion: clusterVersion,
	}
}

// GetVersion returns version information
func (v VersionController) GetVersion(ctx context.Context) *corev2.Version {
	return &corev2.Version{
		Etcd: &etcdVersion.Versions{
			Server:  etcdVersion.Version,
			Cluster: v.clusterVersion,
		},
		SensuBackend: version.Semver(),
	}
}
