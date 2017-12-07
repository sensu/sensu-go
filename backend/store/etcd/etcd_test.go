package etcd

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"testing"

	"github.com/sensu/etcd/clientv3"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEtcd(t *testing.T) {
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
	if e != nil {
		defer func() {
			require.NoError(t, e.Shutdown())
		}()
	}
	assert.NoError(t, err)
	if err != nil {
		_ = pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
		_ = pprof.Lookup("threadcreate").WriteTo(os.Stdout, 1)
		_ = pprof.Lookup("heap").WriteTo(os.Stdout, 1)
		assert.FailNow(t, "unable to start new etcd")
	}

	client, err := e.NewClient()
	assert.NoError(t, err)
	kv := clientv3.NewKV(client)
	assert.NotNil(t, kv)

	putsResp, err := kv.Put(context.Background(), "key", "value")
	assert.NoError(t, err)
	assert.NotNil(t, putsResp)

	if putsResp == nil {
		assert.FailNow(t, "got nil put response from etcd")
	}

	getResp, err := kv.Get(context.Background(), "key")
	assert.NoError(t, err)
	assert.NotNil(t, getResp)

	if getResp == nil {
		assert.FailNow(t, "got nil get response from etcd")
	}
	assert.Equal(t, 1, len(getResp.Kvs))
	assert.Equal(t, "key", string(getResp.Kvs[0].Key))
	assert.Equal(t, "value", string(getResp.Kvs[0].Value))

	require.NoError(t, e.Shutdown())
}
