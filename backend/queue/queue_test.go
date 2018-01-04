package queue

import (
	"context"
	"fmt"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/stretchr/testify/assert"
)

func newEtcd(t *testing.T) (*clientv3.Client, func()) {
	tmpDir, remove := testutil.TempDir(t)

	ports := make([]int, 2)
	err := testutil.RandomPorts(ports)
	if err != nil {
		t.Fatal(err)
	}
	clURL := fmt.Sprintf("http://127.0.0.1:%d", ports[0])
	apURL := fmt.Sprintf("http://127.0.0.1:%d", ports[1])
	initCluster := fmt.Sprintf("default=%s", apURL)

	cfg := etcd.NewConfig()
	cfg.DataDir = tmpDir
	cfg.ListenClientURL = clURL
	cfg.ListenPeerURL = apURL
	cfg.InitialCluster = initCluster
	cfg.InitialClusterState = etcd.ClusterStateNew
	cfg.InitialAdvertisePeerURL = apURL
	cfg.Name = "default"

	e, err := etcd.NewEtcd(cfg)
	assert.NoError(t, err)

	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{clURL},
	})

	assert.NoError(t, err)

	return client, func() {
		if e != nil {
			assert.NoError(t, e.Shutdown())
		}
		remove()
	}
}

func TestEnqueue(t *testing.T) {
	client, cleanup := newEtcd(t)
	defer cleanup()
	queue := New("test", client)
	err := queue.Enqueue(context.Background(), "test value")
	assert.NoError(t, err)

}
