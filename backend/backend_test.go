package backend

import (
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestHTTPListener(t *testing.T) {
	path, remove := testutil.TempDir(t)
	defer remove()

	ports := make([]int, 4)
	err := testutil.RandomPorts(ports)
	if err != nil {
		log.Panic(err)
	}
	clURL := fmt.Sprintf("http://127.0.0.1:%d", ports[0])
	apURL := fmt.Sprintf("http://127.0.0.1:%d", ports[1])
	agentPort := ports[2]
	apiPort := ports[3]
	initCluster := fmt.Sprintf("default=%s", apURL)
	fmt.Println(initCluster)

	b, err := NewBackend(&Config{
		AgentHost:                   "127.0.0.1",
		AgentPort:                   agentPort,
		APIHost:                     "127.0.0.1",
		APIPort:                     apiPort,
		DashboardHost:               "127.0.0.1",
		StateDir:                    path,
		EtcdListenClientURL:         clURL,
		EtcdListenPeerURL:           apURL,
		EtcdInitialCluster:          initCluster,
		EtcdInitialClusterState:     etcd.ClusterStateNew,
		EtcdInitialAdvertisePeerURL: apURL,
	})
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "failed to start backend")
	}

	go func() {
		err = b.Run()
		assert.NoError(t, err)
	}()

	for i := 0; i < 5; i++ {
		conn, derr := net.Dial("tcp", fmt.Sprintf("localhost:%d", agentPort))
		if derr != nil {
			fmt.Println("Waiting for backend to start")
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			conn.Close()
			continue
		}
	}

	client, err := transport.Connect(fmt.Sprintf("ws://localhost:%d/", agentPort))
	assert.NoError(t, err)
	assert.NotNil(t, client)

	msg := &transport.Message{
		Type:    types.AgentHandshakeType,
		Payload: []byte("{}"),
	}
	err = client.Send(msg)
	assert.NoError(t, err)
	resp, err := client.Receive()
	assert.NoError(t, err)
	assert.Equal(t, types.BackendHandshakeType, resp.Type)

	assert.NoError(t, client.Close())
	b.Stop()
}

func TestHTTPSListener(t *testing.T) {
	path, remove := testutil.TempDir(t)
	defer remove()

	ports := make([]int, 5)
	err := testutil.RandomPorts(ports)
	if err != nil {
		log.Panic(err)
	}
	clURL := fmt.Sprintf("https://127.0.0.1:%d", ports[0])
	apURL := fmt.Sprintf("http://127.0.0.1:%d", ports[1])
	agentPort := ports[2]
	apiPort := ports[3]
	dashboardPort := ports[4]
	initCluster := fmt.Sprintf("default=%s", apURL)
	fmt.Println(initCluster)

	b, err := NewBackend(&Config{
		AgentHost:                   "127.0.0.1",
		AgentPort:                   agentPort,
		APIHost:                     "127.0.0.1",
		APIPort:                     apiPort,
		DashboardHost:               "127.0.0.1",
		DashboardPort:               dashboardPort,
		StateDir:                    path,
		EtcdListenClientURL:         clURL,
		EtcdListenPeerURL:           apURL,
		EtcdInitialCluster:          initCluster,
		EtcdInitialClusterState:     etcd.ClusterStateNew,
		EtcdInitialAdvertisePeerURL: apURL,
		TLS: &types.TLSConfig{"../util/ssl/etcd1.pem", "../util/ssl/etcd1-key.pem", "../util/ssl/ca.pem", true},
	})
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "failed to start backend")
	}

	go func() {
		err = b.Run()
		assert.NoError(t, err)
	}()

	for i := 0; i < 5; i++ {
		conn, derr := net.Dial("tcp", fmt.Sprintf("localhost:%d", agentPort))
		if derr != nil {
			fmt.Println("Waiting for agentd to start")
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			conn.Close()
			continue
		}
	}

	for i := 0; i < 5; i++ {
		conn, derr := net.Dial("tcp", fmt.Sprintf("localhost:%d", apiPort))
		if derr != nil {
			fmt.Println("Waiting for apid to start")
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			conn.Close()
			continue
		}
	}

	for i := 0; i < 5; i++ {
		conn, derr := net.Dial("tcp", fmt.Sprintf("localhost:%d", dashboardPort))
		if derr != nil {
			fmt.Println("Waiting for dashboardd to start")
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			conn.Close()
			continue
		}
	}

	client, err := transport.Connect(fmt.Sprintf("ws://localhost:%d/", agentPort))
	assert.NoError(t, err)
	assert.NotNil(t, client)

	msg := &transport.Message{
		Type:    types.AgentHandshakeType,
		Payload: []byte("{}"),
	}
	err = client.Send(msg)
	assert.NoError(t, err)
	resp, err := client.Receive()
	assert.NoError(t, err)
	assert.Equal(t, types.BackendHandshakeType, resp.Type)

	assert.NoError(t, client.Close())
	b.Stop()
}
