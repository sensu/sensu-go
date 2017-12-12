package etcd

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testWithEtcd(t *testing.T, f func(store.Store)) {
	tmpDir, remove := testutil.TempDir(t)
	defer remove()

	ports := make([]int, 2)
	err := testutil.RandomPorts(ports)
	if err != nil {
		log.Panic(err)
	}
	clURL := fmt.Sprintf("http://127.0.0.1:%d", ports[0])
	apURL := fmt.Sprintf("http://127.0.0.1:%d", ports[1])
	initCluster := fmt.Sprintf("default=%s", apURL)

	cfg := NewConfig()
	cfg.DataDir = tmpDir
	cfg.ListenClientURL = clURL
	cfg.ListenPeerURL = apURL
	cfg.InitialCluster = initCluster
	cfg.InitialClusterState = ClusterStateNew
	cfg.InitialAdvertisePeerURL = apURL
	cfg.Name = "default"

	e, err := NewEtcd(cfg)
	assert.NoError(t, err)
	if e != nil {
		defer func() {
			require.NoError(t, e.Shutdown())
		}()
	}

	s, err := e.NewStore()
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "failed to get store from etcd")
	}

	// Mock a default organization
	require.NoError(t, s.UpdateOrganization(context.Background(), types.FixtureOrganization("default")))

	// Mock a default environment
	require.NoError(t, s.UpdateEnvironment(context.Background(), types.FixtureEnvironment("default")))

	f(s)
}
