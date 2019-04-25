package etcd

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/sensu/sensu-go/types"
)

// GetClusterHealth retrieves the cluster health
func (s *Store) GetClusterHealth(ctx context.Context, cluster clientv3.Cluster, etcdClientTLSConfig *tls.Config) *types.HealthResponse {
	healthResponse := &types.HealthResponse{}

	// Do a get op against every cluster member. Collect the  memberIDs and
	// op errors into a response map, and return this map as etcd health
	// information.
	mList, err := cluster.MemberList(ctx)
	if err != nil {
		logger.WithError(err).Warning("could not get the cluster member list")
		return healthResponse
	}
	healthResponse.Header = mList.Header

	for _, member := range mList.Members {
		health := &types.ClusterHealth{
			MemberID: member.ID,
			Name:     member.Name,
		}

		cli, cliErr := clientv3.New(clientv3.Config{
			Endpoints:   member.ClientURLs,
			DialTimeout: 5 * time.Second,
			TLS:         etcdClientTLSConfig,
		})

		if cliErr != nil {
			logger.WithField("member", member.ID).WithError(cliErr).Error("unhealthy cluster member")
			health.Err = cliErr.Error()
			health.Healthy = false
			healthResponse.ClusterHealth = append(healthResponse.ClusterHealth, health)
			continue
		}
		defer func() {
			_ = cli.Close()
		}()

		_, getErr := cli.Get(context.Background(), "health")

		if getErr == nil || getErr == rpctypes.ErrPermissionDenied {
			health.Err = ""
			health.Healthy = true
		} else {
			health.Err = getErr.Error()
			health.Healthy = false
		}

		healthResponse.ClusterHealth = append(healthResponse.ClusterHealth, health)
	}

	alarmResponse, err := s.client.Maintenance.AlarmList(ctx)

	if err != nil {
		logger.WithError(err).Error("failed to fetch etcd alarm list")
	}

	healthResponse.Alarms = alarmResponse.Alarms
	return healthResponse
}
