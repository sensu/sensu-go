package backend

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sensu/sensu-go/backend/resource"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/time/rate"

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
	"github.com/sensu/sensu-go/backend/store/postgres"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	etcdstorev2 "github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sensu/sensu-go/backend/tessend"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/metrics"
	"github.com/sensu/sensu-go/system"
)

var pgWrapper = postgres.NewResourceWrapper(storev2.WrapResource)

func init() {
	// Replace the etcd resource wrapper with our postgres resource wrapper.
	// This is technical debt - it's a package global and not safe to change
	// except under specific conditions.
	//
	// This should go away once we add the postgres configuration store.
	pgWrapper.EnablePostgres()
	storev2.WrapResource = pgWrapper.WrapResource
}

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
	Daemons                []daemon.Daemon
	ConfigStore            storev2.Interface
	EntityStore            store.EntityStore
	EventStore             store.EventStore
	SilenceStore           store.SilenceStore
	KeepaliveStore         storev2.KeepaliveStore
	RingPool               *ringv2.RingPool
	GraphQLService         *graphql.Service
	SecretsProviderManager *secrets.ProviderManager
	HealthRouter           *routers.HealthRouter
	EtcdClientTLSConfig    *tls.Config
	APIDConfig             apid.Config
	PipelineAdapterV1      pipeline.AdapterV1
	LicenseGetter          licensing.Getter
	Bus                    messaging.MessageBus

	Cfg *Config
}

func errorReporter(event pq.ListenerEventType, err error) {
	if err != nil {
		logger.WithError(err).WithField("event", event).Error("postgres notification error")
		return
	}
	switch event {
	case 0:
		logger.Info("postgres NOTIFY listener connected")
	case 1:
		logger.Info("postgres NOTIFY listener disconnected")
	case 2:
		logger.Info("postgres NOTIFY listener reconnected")
	}
}

func initDevModeStateStore(_ context.Context, b *Backend, client *clientv3.Client, _ *Config) {
	store := etcdstore.NewStore(client)
	b.ConfigStore = etcdstorev2.NewStore(client)
	b.KeepaliveStore = storev2.NewLegacyKeepaliveStore(store)
	b.EntityStore = store
	b.SilenceStore = store
	b.RingPool = ringv2.NewRingPool(func(path string) ringv2.Interface {
		return ringv2.New(client, path)
	})
}

func initPGStateStore(ctx context.Context, b *Backend, client *clientv3.Client, db *pgxpool.Pool, config *Config) error {
	eventStore, err := postgres.NewEventStore(db, postgres.NewSilenceStore(db), b.Cfg.Store.PostgresStateStore, 20)
	if err != nil {
		return err
	}
	b.EventStore = eventStore

	entityStore := postgres.NewEntityStore(db)
	b.EntityStore = entityStore

	b.SilenceStore = postgres.NewSilenceStore(db)

	pgStore := postgres.Store{
		EventStore:   eventStore,
		EntityStore:  entityStore,
		SilenceStore: b.SilenceStore,
	}

	b.KeepaliveStore = storev2.NewLegacyKeepaliveStore(pgStore)
	b.EntityStore = entityStore

	// Create the ring pool for round-robin functionality
	// Set up new postgres ringpool
	pgDSN := b.Cfg.Store.PostgresStateStore.DSN
	listener := pq.NewListener(pgDSN, time.Second, time.Minute, errorReporter)
	pgBus := postgres.NewBus(ctx, listener)
	b.RingPool = ringv2.NewRingPool(func(path string) ringv2.Interface {
		ring, err := postgres.NewRing(db, pgBus, path)
		if err != nil {
			logger.WithError(err).Error("error creating round-robin ring")
			logger.Error("round-robin scheduling may be in a degraded state!")
			panic(err)
		}
		return ring
	})
	return nil
}

// Initialize instantiates a Backend struct with the provided config, by
// configuring etcd and establishing a list of daemons, which constitute our
// backend. The daemons will later be started according to their position in the
// b.Daemons list, and stopped in reverse order
func Initialize(ctx context.Context, etcdConfigClient *clientv3.Client, pgConfigDB *pgxpool.Pool, pgStateDB *pgxpool.Pool, config *Config) (*Backend, error) {
	var err error
	// Initialize a Backend struct
	b := &Backend{Cfg: config}

	if config.DevMode {
		initDevModeStateStore(ctx, b, etcdConfigClient, config)
	} else {
		if err := initPGStateStore(ctx, b, etcdConfigClient, pgStateDB, config); err != nil {
			return nil, err
		}
		b.ConfigStore = postgres.NewConfigStore(pgConfigDB)
	}

	var clusterID string
	if clusterID, err = GetClusterID(ctx, b.ConfigStore); err != nil {
		return nil, err
	}

	jwtClient := api.JWT{Store: b.ConfigStore}
	jwtSecret, err := jwtClient.GetSecret(ctx)
	if err != nil {
		return nil, err
	}
	// TODO: don't use global variables
	jwt.SetSecret(jwtSecret)

	backendID := etcd.NewBackendIDGetter(ctx, etcdConfigClient)
	b.Daemons = append(b.Daemons, backendID)

	// Initialize an etcd getter
	queueGetter := queue.EtcdGetter{Client: etcdConfigClient, BackendIDGetter: backendID}

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
	messaging.MultiplexSignal(ctx, bus, syscall.SIGHUP)

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
	assetGetter, err := assetManager.StartAssetManager(ctx, rate.NewLimiter(limit, b.Cfg.AssetsBurstLimit))
	if err != nil {
		return nil, fmt.Errorf("error initializing asset manager: %s", err)
	}

	// Create sensu-system namespace and backend entity
	br := resource.New(b.ConfigStore, bus)
	if err := br.EnsureBackendResources(ctx); err != nil {
		return nil, fmt.Errorf("error creating system namespace and backend entity: %s", err.Error())
	}

	// Initialize the secrets provider manager
	b.SecretsProviderManager = secrets.NewProviderManager(br)

	auth := &rbac.Authorizer{Store: b.ConfigStore}

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
		Store:        b.ConfigStore,
		StoreTimeout: storeTimeout,
	}

	// Initialize PipelineAdapterV1 filter adapters
	legacyFilterAdapter := &filter.LegacyAdapter{
		AssetGetter:  assetGetter,
		Store:        b.ConfigStore,
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
		Store:                  b.ConfigStore,
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
		Store:                  b.ConfigStore,
		StoreTimeout:           storeTimeout,
	}

	b.PipelineAdapterV1.HandlerAdapters = []pipeline.HandlerAdapter{
		legacyHandlerAdapter,
	}

	pipelineDaemon.AddAdapter(&b.PipelineAdapterV1)
	b.Daemons = append(b.Daemons, pipelineDaemon)

	// Initialize eventd
	event, err := eventd.New(
		ctx,
		eventd.Config{
			Store:               b.ConfigStore,
			EventStore:          b.EventStore,
			Bus:                 bus,
			LivenessFactory:     liveness.EtcdFactory(ctx, etcdConfigClient),
			Client:              etcdConfigClient,
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
		ctx,
		schedulerd.Config{
			Store:                  b.ConfigStore,
			Bus:                    bus,
			QueueGetter:            queueGetter,
			RingPool:               b.RingPool,
			Client:                 etcdConfigClient,
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
	entityConfigWatcher := agentd.GetEntityConfigWatcher(ctx, b.ConfigStore)

	// Prepare the etcd etcdConfigClient TLS config
	etcdClientTLSInfo := (transport.TLSInfo)(config.Store.EtcdConfigurationStore.ClientTLSInfo)
	etcdClientTLSConfig, err := etcdClientTLSInfo.ClientConfig()
	if err != nil {
		return nil, err
	}
	b.EtcdClientTLSConfig = etcdClientTLSConfig

	// Initialize keepalived
	keepalive, err := keepalived.New(keepalived.Config{
		Client:                etcdConfigClient,
		DeregistrationHandler: config.DeregistrationHandler,
		Bus:                   bus,
		Store:                 b.ConfigStore,
		EventStore:            b.EventStore,
		KeepaliveStore:        b.KeepaliveStore,
		LivenessFactory:       liveness.EtcdFactory(ctx, etcdConfigClient),
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
		Store:      b.ConfigStore,
	}
	authenticator.AddProvider(provider)

	var clusterVersion string

	if !config.DevMode {
		// get cluster version from first available etcd endpoint
		endpoints := etcdConfigClient.Endpoints()
		for _, ep := range endpoints {
			status, err := etcdConfigClient.Status(ctx, ep)
			if err != nil {
				logger.WithError(err).Error("error getting etcd cluster version info")
				continue
			}
			clusterVersion = status.Version
			break
		}
	} else {
		status, err := etcdConfigClient.Status(ctx, "http://127.0.0.1:2379")
		if err != nil {
			logger.WithError(err).Error("error getting etcd cluster info")
		} else {
			clusterVersion = status.Version
		}
	}

	// Load the JWT key pair
	if err := jwt.LoadKeyPair(viper.GetString(FlagJWTPrivateKeyFile), viper.GetString(FlagJWTPublicKeyFile)); err != nil {
		logger.WithError(err).Error("could not load the key pair for the JWT signature")
	}

	// Initialize the health router
	b.HealthRouter = routers.NewHealthRouter(actions.HealthController{})

	// Initialize GraphQL service
	b.GraphQLService, err = graphql.NewService(graphql.ServiceConfig{
		AssetClient:       api.NewAssetClient(b.ConfigStore, auth),
		CheckClient:       api.NewCheckClient(b.ConfigStore, actions.NewCheckController(b.ConfigStore, queueGetter), auth),
		EntityClient:      api.NewEntityClient(b.EntityStore, b.ConfigStore, b.EventStore, auth),
		EventClient:       api.NewEventClient(b.EventStore, auth, bus),
		EventFilterClient: api.NewEventFilterClient(b.ConfigStore, auth),
		HandlerClient:     api.NewHandlerClient(b.ConfigStore, auth),
		HealthController:  actions.HealthController{},
		MutatorClient:     api.NewMutatorClient(b.ConfigStore, auth),
		SilencedClient:    api.NewSilencedClient(b.SilenceStore, auth),
		NamespaceClient:   api.NewNamespaceClient(b.ConfigStore, auth),
		HookClient:        api.NewHookConfigClient(b.ConfigStore, auth),
		UserClient:        api.NewUserClient(b.ConfigStore, auth),
		RBACClient:        api.NewRBACClient(b.ConfigStore, auth),
		VersionController: actions.NewVersionController(clusterVersion),
		MetricGatherer:    prometheus.DefaultGatherer,
		GenericClient:     &api.GenericClient{Store: b.ConfigStore, Auth: auth},
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing graphql.Service: %s", err)
	}

	// Initialize apid
	b.APIDConfig = apid.Config{
		ListenAddress:       config.APIListenAddress,
		RequestLimit:        config.APIRequestLimit,
		WriteTimeout:        config.APIWriteTimeout,
		URL:                 config.APIURL,
		Bus:                 bus,
		Store:               b.ConfigStore,
		EventStore:          b.EventStore,
		QueueGetter:         queueGetter,
		TLS:                 config.TLS,
		Cluster:             etcdConfigClient.Cluster,
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
		ctx,
		tessend.Config{
			Store:      b.ConfigStore,
			EventStore: b.EventStore,
			RingPool:   b.RingPool,
			Client:     etcdConfigClient,
			Bus:        bus,
			ClusterID:  clusterID,
		})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", tessen.Name(), err)
	}
	b.Daemons = append(b.Daemons, tessen)

	// Initialize agentd
	agent, err := agentd.New(agentd.Config{
		Host:          config.AgentHost,
		Port:          config.AgentPort,
		Bus:           bus,
		Store:         b.ConfigStore,
		TLS:           config.AgentTLSOptions,
		RingPool:      b.RingPool,
		WriteTimeout:  config.AgentWriteTimeout,
		Watcher:       entityConfigWatcher,
		HealthRouter:  b.HealthRouter,
		Authenticator: authenticator,
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", agent.Name(), err)
	}
	b.Daemons = append(b.Daemons, agent)

	return b, nil
}

// Run starts all of the Backend server's daemons
func (b *Backend) Run(ctx context.Context) error {
	var derr error

	eg := errGroup{
		out: make(chan error),
	}

	defer eg.WaitStop()

	// crash the stopgroup after a hard-coded timeout, when etcd is not embedded.
	sg := &stopGroup{crashOnTimeout: true, waitTime: 30 * time.Second}

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
			go metricsBridge.Run(ctx)
		}
	}

	errCtx, errCancel := context.WithCancel(ctx)
	defer errCancel()
	eg.Go(errCtx)

	logger.Warn("backend is running and ready to accept events")

	select {
	case err := <-eg.Err():
		logger.WithError(err).Error("backend stopped working and is restarting")
	case <-ctx.Done():
		logger.Info("backend shutting down")
	}
	if err := sg.Stop(); err != nil {
		if derr == nil {
			derr = err
		}
	}
	if derr == nil && ctx.Err() != context.Canceled {
		derr = ctx.Err()
	}

	return derr
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
	for _, d := range e.daemons {
		d := d
		go func() {
			defer e.wg.Done()
			select {
			case err := <-d.Err():
				err = fmt.Errorf("error from %s: %s", d.Name(), err)
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
