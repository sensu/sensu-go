package backend

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func retryConnect(t *testing.T, address string) {
	var err error
	var conn net.Conn
	for i := 0; i < 5; i++ {
		conn, err = net.Dial("tcp", address)
		if err == nil {
			conn.Close()
			continue
		}
		time.Sleep(time.Duration(i) * time.Second)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestBackendHTTPListener(t *testing.T) {
	// tt = Test Table
	tt := []struct {
		name       string
		httpScheme string
		wsScheme   string
		tls        *types.TLSOptions
	}{
		{"HTTP", "http", "ws", nil},
		{"HTTPS", "https", "wss", &types.TLSOptions{"../util/ssl/etcd1.pem", "../util/ssl/etcd1-key.pem", "../util/ssl/ca.pem", false}},
	}
	// tc = Test Case
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			path, remove := testutil.TempDir(t)
			defer remove()

			ports := make([]int, 5)
			err := testutil.RandomPorts(ports)
			if err != nil {
				log.Panic(err)
			}
			clURL := fmt.Sprintf("%s://127.0.0.1:%d", tc.httpScheme, ports[0])
			apURL := fmt.Sprintf("%s://127.0.0.1:%d", tc.httpScheme, ports[1])
			agentPort := ports[2]
			apiPort := ports[3]
			dashboardPort := ports[4]
			initCluster := fmt.Sprintf("default=%s", apURL)

			tlsOpts := tc.tls

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
				TLS: tlsOpts,
			})
			assert.NoError(t, err)
			if err != nil {
				assert.FailNow(t, "failed to start backend")
			}

			go func() {
				err = b.Run()
				assert.NoError(t, err)
			}()

			retryConnect(t, fmt.Sprintf("localhost:%d", agentPort))
			retryConnect(t, fmt.Sprintf("localhost:%d", apiPort))
			retryConnect(t, fmt.Sprintf("localhost:%d", dashboardPort))

			userCredentials := base64.StdEncoding.EncodeToString([]byte("agent:P@ssw0rd!"))

			hdr := http.Header{
				"Authorization":                  {"Basic " + userCredentials},
				transport.HeaderKeyEnvironment:   {"default"},
				transport.HeaderKeyOrganization:  {"default"},
				transport.HeaderKeyAgentID:       {"agent"},
				transport.HeaderKeySubscriptions: {},
			}
			client, err := transport.Connect(fmt.Sprintf("%s://localhost:%d/", tc.wsScheme, agentPort), tc.tls, hdr)
			assert.NoError(t, err)
			require.NotNil(t, client)

			assert.NoError(t, client.Close())
			b.Stop()
		})
	}
}
