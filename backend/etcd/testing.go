package etcd

import (
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/stretchr/testify/require"
)

// NewTestEtcd creates a new Etcd for testing purposes.
func NewTestEtcd(t *testing.T) (*Etcd, func()) {
	tmpDir, remove := testutil.TempDir(t)

	clURL := "http://127.0.0.1:0"
	apURL := "http://127.0.0.1:0"
	initCluster := fmt.Sprintf("default=%s", apURL)

	cfg := NewConfig()
	cfg.DataDir = tmpDir
	cfg.AdvertiseClientURLs = []string{clURL}
	cfg.ListenClientURLs = []string{clURL}
	cfg.ListenPeerURLs = []string{apURL}
	cfg.InitialCluster = initCluster
	cfg.InitialClusterState = ClusterStateNew
	cfg.InitialAdvertisePeerURLs = []string{apURL}
	cfg.Name = "default"

	e, err := NewEtcd(cfg)
	require.NoError(t, err)
	return e, func() {
		defer remove()
		defer func() {
			require.NoError(t, e.Shutdown())
		}()
	}
}
