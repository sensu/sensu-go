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
	etcdstorev1 "github.com/sensu/sensu-go/backend/store/etcd"
	etcdstorev2 "github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
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
				DevMode:                true,
				DisablePlatformMetrics: true,
				Store: StoreConfig{
					EtcdConfigurationStore: etcdstorev1.Config{
						URLs:          []string{clURL},
						ClientTLSInfo: tlsInfo,
						LogLevel:      "error",
					},
				},
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			client, err := devModeClient(ctx, cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = client.Close() }()

			// note that the pg db is nil, which is fine when DevMode is enabled
			b, err := Initialize(ctx, client, nil, nil, cfg)
			if err != nil {
				t.Fatalf("failed to start backend: %s", err)
			}

			store := etcdstorev2.NewStore(client)
			if err := seeds.SeedInitialDataWithContext(context.Background(), store); err != nil {
				t.Fatal(err)
			}

			var runError error
			var runWg sync.WaitGroup
			runWg.Add(1)
			go func() {
				defer runWg.Done()
				runError = b.Run(ctx)
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
			tclient, _, err := transport.Connect(fmt.Sprintf("%s://127.0.0.1:%d/", tc.wsScheme, agentPort), tc.tls, hdr, 5)
			require.NoError(t, err)
			require.NotNil(t, tclient)

			assert.NoError(t, tclient.Close())
			cancel()
		})
	}
}

func devModeClient(ctx context.Context, config *Config) (*clientv3.Client, error) {
	// Initialize and start etcd, because we'll need to provide an etcd client to
	// the Wizard bus, which requires etcd to be started.
	cfg := etcd.NewConfig()
	cfg.DataDir = config.StateDir
	if urls := config.Store.EtcdConfigurationStore.URLs; len(urls) > 0 {
		cfg.ListenClientURLs = urls
	} else {
		cfg.ListenClientURLs = []string{"http://127.0.0.1:2379"}
	}
	cfg.ListenPeerURLs = []string{"http://127.0.0.1:0"}
	cfg.InitialCluster = "dev=http://127.0.0.1:0"
	cfg.InitialClusterState = "new"
	cfg.InitialAdvertisePeerURLs = cfg.ListenPeerURLs
	cfg.AdvertiseClientURLs = cfg.ListenClientURLs
	cfg.Name = "dev"
	cfg.LogLevel = config.LogLevel
	cfg.ClientLogLevel = config.Store.EtcdConfigurationStore.LogLevel

	// Start etcd
	e, err := etcd.NewEtcd(cfg)
	if err != nil {
		return nil, fmt.Errorf("error starting etcd: %s", err)
	}
	go func() {
		<-ctx.Done()
		if err := e.Shutdown(); err != nil {
			logger.Error(err)
		}
	}()

	// Create an etcd client
	client := e.NewEmbeddedClientWithContext(ctx)
	if _, err := client.Get(ctx, "/sensu.io"); err != nil {
		return nil, err
	}
	return client, nil
}
