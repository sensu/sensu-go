package backend

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/seeds"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
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
			_ = conn.Close()
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
		{"HTTPS", "https", "wss", &types.TLSOptions{
			CertFile:           "../util/ssl/etcd1.pem",
			KeyFile:            "../util/ssl/etcd1-key.pem",
			TrustedCAFile:      "../util/ssl/ca.pem",
			InsecureSkipVerify: false}},
	}
	// tc = Test Case
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			dataPath, remove := testutil.TempDir(t)
			defer remove()

			cachePath, cleanup := testutil.TempDir(t)
			defer cleanup()

			clURL := "http://127.0.0.1:0"

			agentPort := 8081
			apiPort := 8080

			var tlsInfo etcd.TLSInfo
			if tc.tls != nil {
				tlsInfo = etcd.TLSInfo{
					ClientCertAuth: true,
					CertFile:       tc.tls.CertFile,
					KeyFile:        tc.tls.KeyFile,
					TrustedCAFile:  tc.tls.TrustedCAFile,
				}
			}

			cfg := &Config{
				AgentHost:              "127.0.0.1",
				AgentPort:              agentPort,
				APIListenAddress:       fmt.Sprintf("127.0.0.1:%d", apiPort),
				StateDir:               dataPath,
				CacheDir:               cachePath,
				TLS:                    tc.tls,
				EtcdClientURLs:         []string{clURL},
				EtcdClientTLSInfo:      tlsInfo,
				DevMode:                true,
				EtcdClientLogLevel:     "error",
				DisablePlatformMetrics: true,
			}
			ctx, cancel := context.WithCancel(context.Background())
			b, err := Initialize(ctx, cfg)
			if err != nil {
				t.Fatalf("failed to start backend: %s", err)
			}

			store := etcdstore.NewStore(b.Client)
			if err := seeds.SeedInitialDataWithContext(context.Background(), store); err != nil {
				t.Fatal(err)
			}

			var runError error
			var runWg sync.WaitGroup
			runWg.Add(1)
			go func() {
				defer runWg.Done()
				runError = b.Run()
			}()
			defer func() {
				runWg.Wait()
				assert.NoError(t, runError)
			}()

			retryConnect(t, fmt.Sprintf("127.0.0.1:%d", agentPort))
			retryConnect(t, fmt.Sprintf("127.0.0.1:%d", apiPort))

			userCredentials := base64.StdEncoding.EncodeToString([]byte("agent:P@ssw0rd!"))

			hdr := http.Header{
				"Authorization":                  {"Basic " + userCredentials},
				transport.HeaderKeyUser:          {"agent"},
				transport.HeaderKeyNamespace:     {"default"},
				transport.HeaderKeyAgentName:     {"agent"},
				transport.HeaderKeySubscriptions: {},
			}
			time.Sleep(5 * time.Second)
			client, _, err := transport.Connect(fmt.Sprintf("%s://127.0.0.1:%d/", tc.wsScheme, agentPort), tc.tls, hdr, 5)
			require.NoError(t, err)
			require.NotNil(t, client)

			assert.NoError(t, client.Close())
			cancel()
		})
	}
}
