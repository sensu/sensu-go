package etcd

import (
	"context"
	"crypto/tls"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func isEmbeddedClient(clientURLs []string) bool {
	// It is assumed that if any of the client URLs have ':0' as their port,
	// the member is embedded and the client doesn't need to dial.
	for _, url := range clientURLs {
		if strings.HasSuffix(url, ":0") {
			return true
		}
	}
	return false
}

func (s *Store) getHealth(ctx context.Context, id uint64, name string, urls []string, tls *tls.Config) *corev2.ClusterHealth {
	health := &corev2.ClusterHealth{
		MemberID: id,
		Name:     name,
	}

	var cli *clientv3.Client
	var cliErr error

	if isEmbeddedClient(urls) {
		cli = s.client
	} else {
		cli, cliErr = clientv3.New(clientv3.Config{
			Endpoints:   urls,
			DialTimeout: 5 * time.Second,
			TLS:         tls,
		})
	}

	if cliErr != nil {
		logger.WithField("member", id).WithField("name", name).WithError(cliErr).Error("unhealthy cluster member")
		health.Err = cliErr.Error()
		return health
	}
	defer func() {
		_ = cli.Close()
	}()

	_, getErr := cli.Get(ctx, "health")

	if getErr == nil || getErr == rpctypes.ErrPermissionDenied {
		health.Err = ""
		health.Healthy = true
	} else {
		health.Err = getErr.Error()
	}

	return health
}

// GetClusterHealth retrieves the cluster health
func (s *Store) GetClusterHealth(ctx context.Context, cluster clientv3.Cluster, etcdClientTLSConfig *tls.Config) *corev2.HealthResponse {
	healthResponse := &corev2.HealthResponse{}

	var timeout time.Duration
	if val := ctx.Value(store.ContextKeyTimeout); val != nil {
		timeout, _ = val.(time.Duration)
	}

	// Do a get op against every cluster member. Collect the  memberIDs and
	// op errors into a response map, and return this map as etcd health
	// information.
	tctx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		tctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	mList, err := cluster.MemberList(tctx)
	if err != nil {
		logger.WithError(err).Error("could not get the cluster member list")
		healthResponse.ClusterHealth = []*corev2.ClusterHealth{
			{
				Name: "etcd client",
				Err:  fmt.Sprintf("error getting cluster member list: %s", err.Error()),
			},
		}
		return healthResponse
	}
	logger.WithField("members", mList.Members).Info("retrieved cluster members")
	healthResponse.Header = mList.Header

	healths := make(chan *corev2.ClusterHealth, len(mList.Members))
	var wg sync.WaitGroup
	wg.Add(len(mList.Members))

	go func() {
		wg.Wait()
		close(healths)
	}()

	for _, member := range mList.Members {
		go func(id uint64, name string, urls []string) {
			defer wg.Done()
			tctx := ctx
			if timeout > 0 {
				var cancel context.CancelFunc
				tctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}
			healths <- s.getHealth(tctx, id, name, urls, etcdClientTLSConfig)
		}(member.ID, member.Name, member.ClientURLs)
	}

	for health := range healths {
		logger.WithField("health", health).Info("cluster member health")
		healthResponse.ClusterHealth = append(healthResponse.ClusterHealth, health)
	}

	sort.Slice(healthResponse.ClusterHealth, func(i, j int) bool {
		return healthResponse.ClusterHealth[i].Name < healthResponse.ClusterHealth[j].Name
	})

	if timeout > 0 {
		var cancel context.CancelFunc
		tctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	alarmResponse, err := s.client.Maintenance.AlarmList(tctx)
	if err != nil {
		logger.WithError(err).Error("failed to fetch etcd alarm list")
	} else {
		logger.WithField("alarms", len(alarmResponse.Alarms)).Info("cluster alarms")
		healthResponse.Alarms = alarmResponse.Alarms
	}

	return healthResponse
}
