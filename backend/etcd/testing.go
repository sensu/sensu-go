package etcd

import (
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/testing/testutil"
)

// NewTestEtcd creates a new Etcd for testing purposes.
func NewTestEtcd(t testing.TB) (*Etcd, func()) {
	t.Helper()
	return NewTestEtcdWithConfig(t, DefaultEtcdTestConfig(t))
}

// NewTestEtcdWithConfig creates a new Etcd with given config for testing purposes.
func NewTestEtcdWithConfig(t testing.TB, cfg *Config) (*Etcd, func()) {
	t.Helper()
	tmpDir, remove := testutil.TempDir(t)
	cfg.DataDir = tmpDir

	e, err := NewEtcd(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return e, func() {
		defer remove()
		defer e.Shutdown()
	}
}

// DefaultEtcdTestConfig creates a new Config with default values for testing purposes.
func DefaultEtcdTestConfig(t testing.TB) *Config {
	t.Helper()

	clURL := "http://127.0.0.1:0"
	apURL := "http://127.0.0.1:0"
	initCluster := fmt.Sprintf("default=%s", apURL)

	cfg := NewConfig()
	cfg.AdvertiseClientURLs = []string{clURL}
	cfg.ListenClientURLs = []string{clURL}
	cfg.ListenPeerURLs = []string{apURL}
	cfg.InitialCluster = initCluster
	cfg.InitialClusterState = ClusterStateNew
	cfg.InitialAdvertisePeerURLs = []string{apURL}
	cfg.Name = "default"
	cfg.LogLevel = "error"

	return cfg
}
