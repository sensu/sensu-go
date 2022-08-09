package backend

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sensu/sensu-go/backend/resource"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/client/pkg/v3/logutil"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/agentd"
	"github.com/sensu/sensu-go/backend/api"
	"github.com/sensu/sensu-go/backend/apid"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql"
	"github.com/sensu/sensu-go/backend/apid/routers"
	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authentication/providers/basic"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/daemon"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/eventd"
	"github.com/sensu/sensu-go/backend/keepalived"
	"github.com/sensu/sensu-go/backend/licensing"
	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/logging"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/pipeline"
	"github.com/sensu/sensu-go/backend/pipeline/filter"
	"github.com/sensu/sensu-go/backend/pipeline/handler"
	"github.com/sensu/sensu-go/backend/pipeline/mutator"
	"github.com/sensu/sensu-go/backend/pipelined"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/schedulerd"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	etcdstorev2 "github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sensu/sensu-go/backend/tessend"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/metrics"
	"github.com/sensu/sensu-go/system"
	"github.com/sensu/sensu-go/util/retry"
)

type ErrStartup struct {
	Err  error
	Name string
}

func (e ErrStartup) Error() string {
	return fmt.Sprintf("error starting %s: %s", e.Name, e.Err)
}

// SelectedMetrics is a list of Prometheus metric names to use when fetching
// metrics for the metrics logger.
var SelectedMetrics = []string{
	"sensu_go_wizard_bus",
	"sensu_go_event_handler_duration",
	"sensu_go_event_handler_duration_sum",
	"sensu_go_event_handler_duration_count",
	"sensu_go_events_processed",
	"sensu_go_agent_sessions",
	"sensu_go_session_errors",
	"sensu_go_websocket_errors",
	"sensu_go_agentd_event_bytes",
	"sensu_go_agentd_event_bytes_sum",
	"sensu_go_agentd_event_bytes_count",
	"sensu_go_eventd_create_proxy_entity_duration",
	"sensu_go_eventd_create_proxy_entity_duration_sum",
	"sensu_go_eventd_create_proxy_entity_duration_count",
	"sensu_go_eventd_update_event_duration",
	"sensu_go_eventd_update_event_duration_sum",
	"sensu_go_eventd_update_event_duration_count",
	"sensu_go_eventd_bus_publish_duration",
	"sensu_go_eventd_bus_publish_duration_sum",
	"sensu_go_eventd_bus_publish_duration_count",
	"sensu_go_eventd_liveness_factory_duration",
	"sensu_go_eventd_liveness_factory_duration_sum",
	"sensu_go_eventd_liveness_factory_duration_count",
	"sensu_go_eventd_switches_alive_duration",
	"sensu_go_eventd_switches_alive_duration_sum",
	"sensu_go_eventd_switches_alive_duration_count",
	"sensu_go_eventd_switches_bury_duration",
	"sensu_go_eventd_switches_bury_duration_sum",
	"sensu_go_eventd_switches_bury_duration_count",
	"sensu_go_lease_ops",
	"sensu_go_pipelined_message_handler_duration",
	"sensu_go_pipelined_message_handler_duration_sum",
	"sensu_go_pipelined_message_handler_duration_count",
	"sensu_go_pipeline_duration",
	"sensu_go_pipeline_duration_sum",
	"sensu_go_pipeline_duration_count",
	"sensu_go_pipeline_resolve_duration",
	"sensu_go_pipeline_resolve_duration_sum",
	"sensu_go_pipeline_resolve_duration_count",
	"sensu_go_pipeline_filter_duration",
	"sensu_go_pipeline_filter_duration_sum",
	"sensu_go_pipeline_filter_duration_count",
	"sensu_go_pipeline_mutator_duration",
	"sensu_go_pipeline_mutator_duration_sum",
	"sensu_go_pipeline_mutator_duration_count",
	"sensu_go_pipeline_handler_duration",
	"sensu_go_pipeline_handler_duration_sum",
	"sensu_go_pipeline_handler_duration_count",
	"sensu_go_asset_fetch_duration",
	"sensu_go_asset_fetch_duration_sum",
	"sensu_go_asset_fetch_duration_count",
	"sensu_go_asset_expand_duration",
	"sensu_go_asset_expand_duration_sum",
	"sensu_go_asset_expand_duration_count",
	"etcd_debugging_mvcc_keys_total",
	"etcd_debugging_mvcc_delete_total",
	"etcd_debugging_mvcc_put_total",
	"etcd_debugging_mvcc_range_total",
	"etcd_debugging_mvcc_txn_total",
	"etcd_debugging_mvcc_db_total_size_in_bytes",
	"etcd_disk_wal_fsync_duration_seconds_bucket",
	"etcd_disk_backend_commit_duration_seconds_bucket",
	"etcd_network_client_grpc_received_bytes_total",
	"etcd_network_client_grpc_sent_bytes_total",
	"etcd_network_peer_received_bytes_total",
	"etcd_network_peer_sent_bytes_total",
	"etcd_snap_db_fsync_duration_seconds_bucket",
	"graphql_duration_seconds",
	"graphql_duration_seconds_sum",
	"graphql_duration_seconds_count",
}

// Backend represents the backend server, which is used to hold the datastore
// and coordinating the daemons
type Backend struct {
	Client                 *clientv3.Client
	Daemons                []daemon.Daemon
	Etcd                   *etcd.Etcd
	Store                  store.Store
	StoreV2                storev2.Interface
	StoreUpdater           StoreUpdater
	StoreV2Updater         StoreV2Updater
	RingPool               *ringv2.RingPool
	GraphQLService         *graphql.Service
	SecretsProviderManager *secrets.ProviderManager
	HealthRouter           *routers.HealthRouter
	EtcdClientTLSConfig    *tls.Config
	APIDConfig             apid.Config
	PipelineAdapterV1      pipeline.AdapterV1
	LicenseGetter          licensing.Getter
	Bus                    messaging.MessageBus

	ctx       context.Context
	runCtx    context.Context
	runCancel context.CancelFunc
	Cfg       *Config
}

// StoreUpdater offers a way to update an event store to a different
// implementation in-place.
type StoreUpdater interface {
	UpdateStore(to store.Store)
}

// StoreV2Updater is like StoreUpdater, but works on a storev2.Interface.
type StoreV2Updater interface {
	UpdateStore(to storev2.Interface)
}

func newClient(ctx context.Context, config *Config, backend *Backend) (*clientv3.Client, error) {
	if config.NoEmbedEtcd {
		logger.Info("dialing etcd server")
		tlsInfo := (transport.TLSInfo)(config.EtcdClientTLSInfo)
		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return nil, err
		}

		clientURLs := config.EtcdClientURLs
		if len(clientURLs) == 0 {
			clientURLs = config.EtcdAdvertiseClientURLs
		}

		var clientv3Config clientv3.Config

		if config.EtcdClientUsername != "" && config.EtcdClientPassword != "" {
			clientv3Config = clientv3.Config{
				Endpoints:   clientURLs,
				DialTimeout: 5 * time.Second,
				Username:    config.EtcdClientUsername,
				Password:    config.EtcdClientPassword,
				TLS:         tlsConfig,
				DialOptions: []grpc.DialOption{
					grpc.WithReturnConnectionError(),
					grpc.WithBlock(),
				},
			}
		} else {
			clientv3Config = clientv3.Config{
				Endpoints:   clientURLs,
				DialTimeout: 5 * time.Second,
				TLS:         tlsConfig,
				DialOptions: []grpc.DialOption{
					grpc.WithReturnConnectionError(),
					grpc.WithBlock(),
				},
			}
		}

		// Set etcd client log level
		logConfig := logutil.DefaultZapLoggerConfig
		logConfig.Level.SetLevel(etcd.LogLevelToZap(config.EtcdClientLogLevel))
		clientv3Config.LogConfig = &logConfig

		// Don't start up an embedded etcd, return a client that connects to an
		// external etcd instead.
		client, err := clientv3.New(clientv3Config)
		if err != nil {
			return nil, err
		}
		if _, err := client.Get(ctx, "/sensu.io"); err != nil {
			return nil, err
		}
		return client, nil
	}

	// Initialize and start etcd, because we'll need to provide an etcd client to
	// the Wizard bus, which requires etcd to be started.
	cfg := etcd.NewConfig()
	cfg.DataDir = config.StateDir
	cfg.ListenClientURLs = config.EtcdListenClientURLs
	cfg.ListenPeerURLs = config.EtcdListenPeerURLs
	cfg.InitialCluster = config.EtcdInitialCluster
	cfg.InitialClusterState = config.EtcdInitialClusterState
	cfg.InitialAdvertisePeerURLs = config.EtcdInitialAdvertisePeerURLs
	cfg.AdvertiseClientURLs = config.EtcdAdvertiseClientURLs
	cfg.Discovery = config.EtcdDiscovery
	cfg.DiscoverySrv = config.EtcdDiscoverySrv
	cfg.Name = config.EtcdName
	cfg.LogLevel = config.EtcdLogLevel
	cfg.ClientLogLevel = config.EtcdClientLogLevel

	// Heartbeat interval
	if config.EtcdHeartbeatInterval > 0 {
		cfg.TickMs = config.EtcdHeartbeatInterval
	}

	// Election timeout
	if config.EtcdElectionTimeout > 0 {
		cfg.ElectionMs = config.EtcdElectionTimeout
	}

	// Etcd TLS config
	cfg.ClientTLSInfo = config.EtcdClientTLSInfo
	cfg.PeerTLSInfo = config.EtcdPeerTLSInfo
	cfg.CipherSuites = config.EtcdCipherSuites

	if config.EtcdQuotaBackendBytes != 0 {
		cfg.QuotaBackendBytes = config.EtcdQuotaBackendBytes
	}
	if config.EtcdMaxRequestBytes != 0 {
		cfg.MaxRequestBytes = config.EtcdMaxRequestBytes
	}

	// Unsafe config
	cfg.UnsafeNoFsync = config.EtcdUnsafeNoFsync

	// Start etcd
	e, err := etcd.NewEtcd(cfg)
	if err != nil {
		return nil, fmt.Errorf("error starting etcd: %s", err)
	}

	backend.Etcd = e

	// Create an etcd client
	var client *clientv3.Client
	if config.EtcdUseEmbeddedClient {
		client = e.NewEmbeddedClientWithContext(ctx)
	} else {
		cl, err := e.NewClientContext(backend.runCtx)
		if err != nil {
			return nil, err
		}
		client = cl
	}
	if _, err := client.Get(ctx, "/sensu.io"); err != nil {
		return nil, err
	}
	return client, nil
}

// Initialize instantiates a Backend struct with the provided config, by
// configuring etcd and establishing a list of daemons, which constitute our
// backend. The daemons will later be started according to their position in the
// b.Daemons list, and stopped in reverse order
func Initialize(ctx context.Context, config *Config) (*Backend, error) {
	var err error
	// Initialize a Backend struct
	b := &Backend{Cfg: config}

	b.ctx = ctx
	b.runCtx, b.runCancel = context.WithCancel(b.ctx)

	b.Client, err = newClient(b.RunContext(), config, b)
	if err != nil {
		return nil, err
	}

	// Create the store, which lives on top of etcd
	stor := etcdstore.NewStore(b.Client, config.EtcdName)
	b.Store = stor
	storv2 := etcdstorev2.NewStore(b.Client)
	var storev2Proxy storev2.Proxy
	storev2Proxy.UpdateStore(storv2)
	b.StoreV2 = &storev2Proxy
	b.StoreV2Updater = &storev2Proxy

	// Create the ring pool for round-robin functionality
	b.RingPool = ringv2.NewRingPool(func(path string) ringv2.Interface {
		return ringv2.New(b.Client, path)
	})

	if _, err := stor.GetClusterID(b.RunContext()); err != nil {
		return nil, err
	}

	// Initialize the JWT secret. This method is idempotent and needs to be ran
	// at every startup so the JWT signatures remain valid
	if err := jwt.InitSecret(b.Store); err != nil {
		return nil, err
	}

	storeProxy := store.NewStoreProxy(stor)
	b.StoreUpdater = storeProxy
	b.Store = storeProxy

	logger.Debug("Registering backend...")

	backendID := etcd.NewBackendIDGetter(b.RunContext(), b.Client)
	logger.Debug("Done registering backend.")
	b.Daemons = append(b.Daemons, backendID)

	// Initialize an etcd getter
	queueGetter := queue.EtcdGetter{Client: b.Client, BackendIDGetter: backendID}

	// Initialize the LicenseGetter
	b.LicenseGetter = config.LicenseGetter

	// Initialize the bus
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", bus.Name(), err)
	}
	b.Bus = bus
	b.Daemons = append(b.Daemons, bus)

	// Publish all SIGHUP signals to wizard bus until the provided context is cancelled
	messaging.MultiplexSignal(b.RunContext(), bus, syscall.SIGHUP)

	// Initialize asset manager
	backendEntity := b.getBackendEntity(config)
	logger.WithField("entity", backendEntity).Info("backend entity information")
	var trustedCAFile string
	if config.TLS != nil {
		trustedCAFile = config.TLS.TrustedCAFile
	}
	assetManager := asset.NewManager(config.CacheDir, trustedCAFile, backendEntity, &sync.WaitGroup{})
	limit := b.Cfg.AssetsRateLimit
	if limit == 0 {
		limit = asset.DefaultAssetsRateLimit
	}
	assetGetter, err := assetManager.StartAssetManager(b.RunContext(), rate.NewLimiter(limit, b.Cfg.AssetsBurstLimit))
	if err != nil {
		return nil, fmt.Errorf("error initializing asset manager: %s", err)
	}

	// Create sensu-system namespace and backend entity
	br := resource.New(b.Store, b.Store, bus)
	if err := br.EnsureBackendResources(ctx); err != nil {
		return nil, fmt.Errorf("error creating system namespace and backend entity: %s", err.Error())
	}

	// Initialize the secrets provider manager
	b.SecretsProviderManager = secrets.NewProviderManager(br)

	auth := &rbac.Authorizer{Store: b.Store}

	// Initialize pipelined
	pipelineDaemon, err := pipelined.New(pipelined.Config{
		Bus:         bus,
		BufferSize:  viper.GetInt(FlagPipelinedBufferSize),
		WorkerCount: viper.GetInt(FlagPipelinedWorkers),
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", pipelineDaemon.Name(), err)
	}

	// Initialize PipelineAdapterV1
	storeTimeout := 2 * time.Minute
	b.PipelineAdapterV1 = pipeline.AdapterV1{
		Store:        b.Store,
		StoreTimeout: storeTimeout,
	}

	// Initialize PipelineAdapterV1 filter adapters
	legacyFilterAdapter := &filter.LegacyAdapter{
		AssetGetter:  assetGetter,
		Store:        b.Store,
		StoreTimeout: storeTimeout,
	}
	hasMetricsFilterAdapter := &filter.HasMetricsAdapter{}
	isIncidentFilterAdapter := &filter.IsIncidentAdapter{}
	notSilencedFilterAdapter := &filter.NotSilencedAdapter{}

	b.PipelineAdapterV1.FilterAdapters = []pipeline.FilterAdapter{
		legacyFilterAdapter,
		hasMetricsFilterAdapter,
		isIncidentFilterAdapter,
		notSilencedFilterAdapter,
	}

	// Initialize PipelineAdapterV1 mutator adapters
	legacyMutatorAdapter := &mutator.LegacyAdapter{
		AssetGetter:            assetGetter,
		Executor:               command.NewExecutor(),
		SecretsProviderManager: b.SecretsProviderManager,
		Store:                  b.Store,
		StoreTimeout:           storeTimeout,
	}
	onlyCheckOutputMutatorAdapter := &mutator.OnlyCheckOutputAdapter{}
	jsonMutatorAdapter := &mutator.JSONAdapter{}

	b.PipelineAdapterV1.MutatorAdapters = []pipeline.MutatorAdapter{
		legacyMutatorAdapter,
		onlyCheckOutputMutatorAdapter,
		jsonMutatorAdapter,
	}

	// Initialize PipelineAdapterV1 handler adapters
	legacyHandlerAdapter := &handler.LegacyAdapter{
		AssetGetter:            assetGetter,
		Executor:               command.NewExecutor(),
		LicenseGetter:          b.LicenseGetter,
		SecretsProviderManager: b.SecretsProviderManager,
		Store:                  b.Store,
		StoreTimeout:           storeTimeout,
	}

	b.PipelineAdapterV1.HandlerAdapters = []pipeline.HandlerAdapter{
		legacyHandlerAdapter,
	}

	pipelineDaemon.AddAdapter(&b.PipelineAdapterV1)
	b.Daemons = append(b.Daemons, pipelineDaemon)

	// Initialize eventd
	event, err := eventd.New(
		b.RunContext(),
		eventd.Config{
			Store:               b.StoreV2,
			EventStore:          b.Store,
			Bus:                 bus,
			LivenessFactory:     liveness.EtcdFactory(b.RunContext(), b.Client),
			Client:              b.Client,
			BufferSize:          viper.GetInt(FlagEventdBufferSize),
			WorkerCount:         viper.GetInt(FlagEventdWorkers),
			StoreTimeout:        2 * time.Minute,
			LogPath:             b.Cfg.EventLogFile,
			LogBufferSize:       b.Cfg.EventLogBufferSize,
			LogBufferWait:       b.Cfg.EventLogBufferWait,
			LogParallelEncoders: b.Cfg.EventLogParallelEncoders,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", event.Name(), err)
	}
	b.Daemons = append(b.Daemons, event)

	// Initialize schedulerd
	scheduler, err := schedulerd.New(
		b.RunContext(),
		schedulerd.Config{
			Store:                  b.Store,
			Bus:                    bus,
			QueueGetter:            queueGetter,
			RingPool:               b.RingPool,
			Client:                 b.Client,
			SecretsProviderManager: b.SecretsProviderManager,
		})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", scheduler.Name(), err)
	}
	b.Daemons = append(b.Daemons, scheduler)

	// Use the common TLS flags for agentd if wasn't explicitely configured with
	// its own TLS configuration
	if config.TLS != nil && config.AgentTLSOptions == nil {
		config.AgentTLSOptions = config.TLS
	}

	// Start the entity config watcher, so agentd sessions are notified of updates
	entityConfigWatcher := agentd.GetEntityConfigWatcher(b.ctx, b.Client)

	// Prepare the etcd client TLS config
	etcdClientTLSInfo := (transport.TLSInfo)(config.EtcdClientTLSInfo)
	etcdClientTLSConfig, err := etcdClientTLSInfo.ClientConfig()
	if err != nil {
		return nil, err
	}
	b.EtcdClientTLSConfig = etcdClientTLSConfig

	// Initialize keepalived
	keepalive, err := keepalived.New(keepalived.Config{
		DeregistrationHandler: config.DeregistrationHandler,
		Bus:                   bus,
		Store:                 b.Store,
		StoreV2:               b.StoreV2,
		EventStore:            b.Store,
		LivenessFactory:       liveness.EtcdFactory(b.RunContext(), b.Client),
		RingPool:              b.RingPool,
		BufferSize:            viper.GetInt(FlagKeepalivedBufferSize),
		WorkerCount:           viper.GetInt(FlagKeepalivedWorkers),
		StoreTimeout:          2 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", keepalive.Name(), err)
	}
	b.Daemons = append(b.Daemons, keepalive)

	// Prepare the authentication providers
	authenticator := &authentication.Authenticator{}
	provider := &basic.Provider{
		ObjectMeta: corev2.ObjectMeta{Name: basic.Type},
		Store:      b.Store,
	}
	authenticator.AddProvider(provider)

	var clusterVersion string
	if config.NoEmbedEtcd {
		// get cluster version from first available etcd endpoint
		endpoints := b.Client.Endpoints()
		for _, ep := range endpoints {
			status, err := b.Client.Status(ctx, ep)
			if err != nil {
				logger.WithError(err).Error("error getting etcd cluster version info")
				continue
			}
			clusterVersion = status.Version
			break
		}
	} else {
		clusterVersion = b.Etcd.GetClusterVersion()
	}

	// Load the JWT key pair
	if err := jwt.LoadKeyPair(viper.GetString(FlagJWTPrivateKeyFile), viper.GetString(FlagJWTPublicKeyFile)); err != nil {
		logger.WithError(err).Error("could not load the key pair for the JWT signature")
	}

	// Initialize the health router
	b.HealthRouter = routers.NewHealthRouter(actions.NewHealthController(b.Store, b.Client.Cluster, b.EtcdClientTLSConfig))

	// Initialize GraphQL service
	b.GraphQLService, err = graphql.NewService(graphql.ServiceConfig{
		AssetClient:       api.NewAssetClient(b.Store, auth),
		CheckClient:       api.NewCheckClient(b.Store, actions.NewCheckController(b.Store, queueGetter), auth),
		EntityClient:      api.NewEntityClient(b.Store, b.StoreV2, b.Store, auth),
		EventClient:       api.NewEventClient(b.Store, auth, bus),
		EventFilterClient: api.NewEventFilterClient(b.Store, auth),
		HandlerClient:     api.NewHandlerClient(b.Store, auth),
		HealthController:  actions.NewHealthController(b.Store, b.Client.Cluster, etcdClientTLSConfig),
		MutatorClient:     api.NewMutatorClient(b.Store, auth),
		SilencedClient:    api.NewSilencedClient(b.Store, auth),
		NamespaceClient:   api.NewNamespaceClient(b.Store, b.Store, auth, b.StoreV2),
		HookClient:        api.NewHookConfigClient(b.Store, auth),
		UserClient:        api.NewUserClient(b.Store, auth),
		RBACClient:        api.NewRBACClient(b.Store, auth),
		VersionController: actions.NewVersionController(clusterVersion),
		MetricGatherer:    prometheus.DefaultGatherer,
		GenericClient:     &api.GenericClient{Store: b.Store, Auth: auth},
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing graphql.Service: %s", err)
	}

	// Initialize apid
	b.APIDConfig = apid.Config{
		ListenAddress:       config.APIListenAddress,
		RequestLimit:        config.APIRequestLimit,
		WriteTimeout:        config.APIWriteTimeout,
		ServeWaitTime:       config.APIServeWaitTime,
		URL:                 config.APIURL,
		Bus:                 bus,
		Store:               b.Store,
		Storev2:             b.StoreV2,
		EventStore:          b.Store,
		QueueGetter:         queueGetter,
		TLS:                 config.TLS,
		Cluster:             b.Client.Cluster,
		EtcdClientTLSConfig: etcdClientTLSConfig,
		Authenticator:       authenticator,
		ClusterVersion:      clusterVersion,
		GraphQLService:      b.GraphQLService,
		HealthRouter:        b.HealthRouter,
	}
	newApi, err := apid.New(b.APIDConfig)
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", newApi.Name(), err)
	}
	b.Daemons = append(b.Daemons, newApi)

	// Initialize tessend
	tessen, err := tessend.New(
		b.RunContext(),
		tessend.Config{
			Store:      b.Store,
			EventStore: b.Store,
			RingPool:   b.RingPool,
			Client:     b.Client,
			Bus:        bus,
		})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", tessen.Name(), err)
	}
	b.Daemons = append(b.Daemons, tessen)

	// Initialize agentd
	agent, err := agentd.New(agentd.Config{
		Host:                config.AgentHost,
		Port:                config.AgentPort,
		Bus:                 bus,
		Store:               b.Store,
		TLS:                 config.AgentTLSOptions,
		RingPool:            b.RingPool,
		WriteTimeout:        config.AgentWriteTimeout,
		ServeWaitTime:       config.AgentServeWaitTime,
		Client:              b.Client,
		Watcher:             entityConfigWatcher,
		EtcdClientTLSConfig: b.EtcdClientTLSConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", agent.Name(), err)
	}
	b.Daemons = append(b.Daemons, agent)

	return b, nil
}

func (b *Backend) runOnce() error {
	eCloser := b.StoreUpdater.(closer)
	defer func() { _ = eCloser.Close() }()

	var derr error

	eg := errGroup{
		out: make(chan error),
	}

	defer eg.WaitStop()

	if b.Etcd != nil {
		defer func() {
			logger.Info("shutting down etcd")
			if err := recover(); err != nil {
				trace := string(debug.Stack())
				logger.WithField("panic", trace).WithError(fmt.Errorf("%s", err)).
					Error("recovering from panic due to error, shutting down etcd")
			}
			err := b.Etcd.Shutdown()
			if derr == nil {
				derr = err
			}
		}()
	}

	// crash the stopgroup after a hard-coded timeout, when etcd is not embedded.
	sg := &stopGroup{crashOnTimeout: b.Cfg.NoEmbedEtcd, waitTime: 30 * time.Second}

	// Loop across the daemons in order to start them, then add them to our groups
	for _, d := range b.Daemons {
		logger.Infof("starting daemon: %s", d.Name())
		if err := d.Start(); err != nil {
			_ = sg.Stop()
			return ErrStartup{Err: err, Name: d.Name()}
		}

		// Add the daemon to our errGroup
		eg.daemons = append(eg.daemons, d)

		// Add the daemon to our stopGroup
		sg.Add(d)
	}

	if !b.Cfg.DisablePlatformMetrics {
		consumer := fmt.Sprintf("filelogger://%s", b.Cfg.PlatformMetricsLogFile)
		sighup := make(messaging.ChanSubscriber, 1)
		defer close(sighup)
		subscription, err := b.Bus.Subscribe(messaging.SignalTopic(syscall.SIGHUP), consumer, sighup)
		if err != nil {
			logger.WithError(err).Error("unable to subscribe to SIGHUP signal notifications")
			return err
		}
		defer func() {
			_ = subscription.Cancel()
		}()
		metricsLogWriter, err := logging.NewRotateWriter(b.Cfg.PlatformMetricsLogFile, sighup)
		if err != nil {
			logger.WithError(err).Error("unable to open platform metrics log file")
		} else {
			defer func() { _ = metricsLogWriter.Close() }()
			metricsBridge, err := metrics.NewInfluxBridge(&metrics.InfluxBridgeConfig{
				Writer:      metricsLogWriter,
				Interval:    b.Cfg.PlatformMetricsLoggingInterval,
				Gatherer:    prometheus.DefaultGatherer,
				ErrLogger:   logger,
				Select:      SelectedMetrics,
				ExtraLabels: map[string]string{"backend": getDefaultBackendID()},
			})
			if err != nil {
				logger.WithError(err).Error("unable to start the platform metrics bridge")
				return err
			}
			go metricsBridge.Run(b.RunContext())
		}
	}

	if b.Etcd != nil {
		// Add etcd to our errGroup, since it's not included in the daemon list
		eg.daemons = append(eg.daemons, b.Etcd)
	}

	errCtx, errCancel := context.WithCancel(b.RunContext())
	defer errCancel()
	eg.Go(errCtx)

	logger.Warn("backend is running and ready to accept events")

	select {
	case err := <-eg.Err():
		logger.WithError(err).Error("backend stopped working and is restarting")
	case <-b.RunContext().Done():
		logger.Info("backend shutting down")
	}
	if err := sg.Stop(); err != nil {
		if derr == nil {
			derr = err
		}
	}
	if derr == nil {
		derr = b.RunContext().Err()
	}

	return derr
}

type closer interface {
	Close() error
}

// RunContext returns the context for the current run of the backend.
func (b *Backend) RunContext() context.Context {
	return b.runCtx
}

// RunWithInitializer is like Run but accepts an initialization function to use
// for initialization, instead of using the default Initialize().
func (b *Backend) RunWithInitializer(initialize func(context.Context, *Config) (*Backend, error)) error {
	// we allow inErrChan to leak to avoid panics from other
	// goroutines writing errors to either after shutdown has been initiated.
	backoff := retry.ExponentialBackoff{
		Ctx:                  b.ctx,
		InitialDelayInterval: time.Second,
		MaxDelayInterval:     time.Second,
		Multiplier:           1,
	}

	sighup := make(chan os.Signal, 1)
	signal.Notify(sighup, syscall.SIGHUP)

	err := backoff.Retry(func(int) (bool, error) {
		err := b.runOnce()
		b.Stop()
		if err != nil {
			if b.ctx.Err() != nil {
				logger.Warn("shutting down")
				return true, b.ctx.Err()
			}
			logger.Error(err)
			if _, ok := err.(ErrStartup); ok {
				return true, err
			}
		}

		_ = b.Client.Close()

		// Yes, two levels of retry... this could improve. Unfortunately Initialize()
		// is called elsewhere.
		err = backoff.Retry(func(int) (bool, error) {
			backend, err := initialize(b.ctx, b.Cfg)
			if err != nil && err != context.Canceled {
				logger.Error(err)
				return false, nil
			} else if err == context.Canceled {
				return true, err
			}
			// Replace b with a new backend - this is done to ensure that there is
			// no side effects from the execution of b that have carried over
			b = backend
			return true, nil
		})
		if err != nil {
			return true, err
		}
		return false, nil
	})

	if err == context.Canceled {
		return nil
	}

	return err
}

// Run starts all of the Backend server's daemons
func (b *Backend) Run() error {
	return b.RunWithInitializer(Initialize)
}

type stopper interface {
	Stop() error
	Name() string
}

type stopGroup struct {
	stoppers       []stopper
	crashOnTimeout bool
	waitTime       time.Duration
}

func (s *stopGroup) Add(stopper stopper) {
	s.stoppers = append(s.stoppers, stopper)
}

func (s stopGroup) Stop() (err error) {
	// Reverse the order of our stopGroup so daemons are stopped in the proper
	// order (last one started is first one stopped)
	stoppers := make([]stopper, len(s.stoppers))
	copy(stoppers, s.stoppers)
	for i := len(stoppers)/2 - 1; i >= 0; i-- {
		opp := len(stoppers) - 1 - i
		stoppers[i], stoppers[opp] = stoppers[opp], stoppers[i]
	}

	for _, stpr := range stoppers {
		func() {
			logger.Info("shutting down ", stpr.Name())
			if s.crashOnTimeout {
				ctx, cancel := context.WithTimeout(context.Background(), s.waitTime)
				defer cancel()
				go func() {
					<-ctx.Done()
					if ctx.Err() == context.DeadlineExceeded {
						panic(fmt.Sprintf("%s did not stop within %s", stpr.Name(), s.waitTime))
					}
				}()
			}
			e := stpr.Stop()
			if err == nil {
				err = e
			}
		}()
	}
	return err
}

type errorer interface {
	Err() <-chan error
	Name() string
}

type errGroup struct {
	out     chan error
	daemons []errorer
	wg      sync.WaitGroup
}

func (e *errGroup) Go(ctx context.Context) {
	e.wg.Add(len(e.daemons))
	for _, daemon := range e.daemons {
		daemon := daemon
		go func() {
			defer e.wg.Done()
			select {
			case err := <-daemon.Err():
				err = fmt.Errorf("error from %s: %s", daemon.Name(), err)
				select {
				case e.out <- err:
				case <-ctx.Done():
				}
			case <-ctx.Done():
			}
		}()
	}
}

func (e *errGroup) Err() <-chan error {
	return e.out
}

func (e *errGroup) WaitStop() {
	e.wg.Wait()
}

// Stop the Backend cleanly.
func (b *Backend) Stop() {
	b.runCancel()
}

func (b *Backend) getBackendEntity(config *Config) *corev2.Entity {
	entity := &corev2.Entity{
		EntityClass: corev2.EntityBackendClass,
		System:      getSystemInfo(),
		ObjectMeta: corev2.ObjectMeta{
			Name:        getDefaultBackendID(),
			Labels:      b.Cfg.Labels,
			Annotations: b.Cfg.Annotations,
		},
	}

	if config.DeregistrationHandler != "" {
		entity.Deregistration = corev2.Deregistration{
			Handler: config.DeregistrationHandler,
		}
	}

	return entity
}

// getDefaultBackendID returns the default backend ID
func getDefaultBackendID() string {
	defaultBackendID, err := os.Hostname()
	if err != nil {
		logger.WithError(err).Error("error getting hostname")
		defaultBackendID = "unidentified-sensu-backend"
	}
	return defaultBackendID
}

// getSystemInfo returns the system info of the backend
func getSystemInfo() corev2.System {
	info, err := system.Info()
	if err != nil {
		logger.WithError(err).Error("error getting system info")
	}
	return info
}
