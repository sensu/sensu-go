package etcd

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/sensu/sensu-go/types"
)

func (s *Store) GetClusterHealth(ctx context.Context) []*types.ClusterHealth {
	var healthList []*types.ClusterHealth

	// Do a get op against every cluster member. Collect the  memberIDs and
	// op errors into a response map, and return this map as etcd health
	// information.
	mList, err := s.client.MemberList(context.Background())
	if err != nil {
		return healthList
	}

	for _, member := range mList.Members {
		health := &types.ClusterHealth{
			MemberID: member.ID,
			Name:     member.Name,
		}

		cli, cliErr := clientv3.New(clientv3.Config{
			Endpoints:   member.ClientURLs,
			DialTimeout: 5 * time.Second,
		})

		if err != nil || cli == nil {
			health.Err = cliErr
			health.Healthy = false
			healthList = append(healthList, health)
			continue
		}
		defer cli.Close()

		_, getErr := cli.Get(context.Background(), "health")

		if getErr == nil || getErr == rpctypes.ErrPermissionDenied {
			health.Err = nil
			health.Healthy = true
		} else {
			health.Err = getErr
			health.Healthy = false
		}

		healthList = append(healthList, health)
	}
	return healthList
}
