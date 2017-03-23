package e2e

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/testing/util"
	"github.com/stretchr/testify/assert"
)

func TestAgentKeepalives(t *testing.T) {
	ports := make([]int, 3)
	err := util.RandomPorts(ports)
	if err != nil {
		log.Fatal(err)
	}

	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensu")
	if err != nil {
		log.Panic(err)
	}
	defer os.RemoveAll(tmpDir)

	etcdClientURL := fmt.Sprintf("http://127.0.0.1:%d", ports[0])
	etcdPeerURL := fmt.Sprintf("http://127.0.0.1:%d", ports[1])
	backendPort := ports[2]
	backendWSURL := fmt.Sprintf("ws://127.0.0.1:%d/agents/ws", backendPort)
	backendHTTPURL := fmt.Sprintf("http://127.0.0.1:%d", backendPort)
	initialCluster := fmt.Sprintf("default=%s", etcdPeerURL)

	bep := &backendProcess{
		BackendPort:        backendPort,
		StateDir:           tmpDir,
		EtcdClientURL:      etcdClientURL,
		EtcdPeerURL:        etcdPeerURL,
		EtcdInitialCluster: initialCluster,
	}

	err = bep.Start()
	if err != nil {
		log.Panic(err)
	}

	ap := &agentProcess{
		BackendURL: backendWSURL,
	}

	backendHealthy := false
	for i := 0; i < 10; i++ {
		resp, getErr := http.Get(fmt.Sprintf("%s/health", backendHTTPURL))
		if getErr != nil {
			log.Println("backend not ready, sleeping...")
			time.Sleep(1 * time.Second)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("backend returned non-200 status code: %d\n", resp.StatusCode)
			time.Sleep(1 * time.Second)
			continue
		}
		backendHealthy = true
	}

	err = ap.Start()
	if err != nil {
		log.Panic(err)
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdClientURL},
		DialTimeout: 5 * time.Second,
	})
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "unable to connect to etcd")
	}
	defer client.Close()
	kvc := clientv3.NewKV(client)

	// Give it a second to make sure we've sent a keepalive.
	time.Sleep(1 * time.Second)
	resp, err := kvc.Get(context.Background(), "/sensu.io", clientv3.WithPrefix())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(resp.Kvs))
	bep.Kill()
	ap.Kill()

	assert.True(t, backendHealthy)

	// TODO(greg): Figure out if there's a way for us to only print logs if the
	// test fails.
	b, err := ioutil.ReadAll(bep.Stderr)
	if err != nil {
		log.Panic(err)
	}
	fmt.Print(string(b))

	b, err = ioutil.ReadAll(ap.Stderr)
	if err != nil {
		log.Panic(err)
	}
	fmt.Print(string(b))
}
