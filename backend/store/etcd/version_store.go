package etcd

import (
	"context"

	"github.com/coreos/etcd/client"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/version"
)

// GetVersion gets the current etcd server/cluster version, and the sensu-backend version.
func (s *Store) GetVersion(ctx context.Context, client client.Client) (*corev2.Version, error) {
	etcd, err := client.GetVersion(ctx)
	if err != nil {
		return &corev2.Version{}, err
	}
	version := &corev2.Version{
		Etcd:         etcd,
		SensuBackend: version.Semver(),
	}
	return version, nil
}
