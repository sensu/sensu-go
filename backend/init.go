package backend

import (
	"context"
	"fmt"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
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
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/eventd"
	"github.com/sensu/sensu-go/backend/keepalived"
	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/pipeline"
	"github.com/sensu/sensu-go/backend/pipeline/filter"
	"github.com/sensu/sensu-go/backend/pipeline/handler"
	"github.com/sensu/sensu-go/backend/pipeline/mutator"
	"github.com/sensu/sensu-go/backend/pipelined"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/backend/resource"
	"github.com/sensu/sensu-go/backend/schedulerd"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/tessend"
	"github.com/sensu/sensu-go/command"
)

func InitializeStore(ctx context.Context, client *clientv3.Client, db *pgxpool.Pool, config *Config) (*Backend, error) {
	var err error
	// Initialize a Backend struct
	b := &Backend{Cfg: config}

	if config.DevMode {
		initDevModeStateStore(ctx, b, client, config)
	} else {
		if err := initPGStateStore(ctx, b, client, db, config); err != nil {
			return nil, err
		}
	}

	if _, err := b.Store.GetClusterID(ctx); err != nil {
		return nil, err
	}

	// Initialize the JWT secret. This method is idempotent and needs to be ran
	// at every startup so the JWT signatures remain valid
	if err := jwt.InitSecret(b.Store); err != nil {
		return nil, err
	}

	backendID := etcd.NewBackendIDGetter(ctx, client)
	b.Daemons = append(b.Daemons, backendID)

	// Initialize an etcd getter
	queueGetter := queue.EtcdGetter{Client: client, BackendIDGetter: backendID}

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
	br := resource.New(b.StoreV2, bus)
	if err := br.EnsureBackendResources(ctx); err != nil {
		return nil, fmt.Errorf("error creating system namespace and backend entity: %s", err.Error())
	}

	// Initialize the secrets provider manager
	b.SecretsProviderManager = secrets.NewProviderManager(br)

	auth := &rbac.Authorizer{Store: b.StoreV2}

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
		Store:        b.StoreV2,
		StoreTimeout: storeTimeout,
	}

	// Initialize PipelineAdapterV1 filter adapters
	legacyFilterAdapter := &filter.LegacyAdapter{
		AssetGetter:  assetGetter,
		Store:        b.StoreV2,
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
		Store:                  b.StoreV2,
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
		Store:                  b.StoreV2,
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
			Store:               b.StoreV2,
			EventStore:          b.EventStore,
			Bus:                 bus,
			LivenessFactory:     liveness.EtcdFactory(ctx, client),
			Client:              client,
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
			Store:                  b.StoreV2,
			Bus:                    bus,
			QueueGetter:            queueGetter,
			RingPool:               b.RingPool,
			Client:                 client,
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
	entityConfigWatcher := agentd.GetEntityConfigWatcher(ctx, b.StoreV2)

	// Prepare the etcd client TLS config
	etcdClientTLSInfo := (transport.TLSInfo)(config.Store.EtcdConfigurationStore.ClientTLSInfo)
	etcdClientTLSConfig, err := etcdClientTLSInfo.ClientConfig()
	if err != nil {
		return nil, err
	}
	b.EtcdClientTLSConfig = etcdClientTLSConfig

	// Initialize keepalived
	keepalive, err := keepalived.New(keepalived.Config{
		Client:                client,
		DeregistrationHandler: config.DeregistrationHandler,
		Bus:                   bus,
		Store:                 b.StoreV2,
		EventStore:            b.EventStore,
		KeepaliveStore:        b.KeepaliveStore,
		LivenessFactory:       liveness.EtcdFactory(ctx, client),
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
		Store:      b.StoreV2,
	}
	authenticator.AddProvider(provider)

	var clusterVersion string

	if !config.DevMode {
		// get cluster version from first available etcd endpoint
		endpoints := client.Endpoints()
		for _, ep := range endpoints {
			status, err := client.Status(ctx, ep)
			if err != nil {
				logger.WithError(err).Error("error getting etcd cluster version info")
				continue
			}
			clusterVersion = status.Version
			break
		}
	} else {
		status, err := client.Status(ctx, "http://127.0.0.1:2379")
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
	b.HealthRouter = routers.NewHealthRouter(actions.NewHealthController(b.Store, client.Cluster, b.EtcdClientTLSConfig))

	// Initialize GraphQL service
	b.GraphQLService, err = graphql.NewService(graphql.ServiceConfig{
		AssetClient:       api.NewAssetClient(b.StoreV2, auth),
		CheckClient:       api.NewCheckClient(b.Store, actions.NewCheckController(b.StoreV2, queueGetter), auth),
		EntityClient:      api.NewEntityClient(b.Store, b.StoreV2, b.Store, auth),
		EventClient:       api.NewEventClient(b.Store, auth, bus),
		EventFilterClient: api.NewEventFilterClient(b.StoreV2, auth),
		HandlerClient:     api.NewHandlerClient(b.StoreV2, auth),
		HealthController:  actions.NewHealthController(b.Store, client.Cluster, etcdClientTLSConfig),
		MutatorClient:     api.NewMutatorClient(b.StoreV2, auth),
		SilencedClient:    api.NewSilencedClient(b.Store, auth),
		NamespaceClient:   api.NewNamespaceClient(b.StoreV2, auth),
		HookClient:        api.NewHookConfigClient(b.StoreV2, auth),
		UserClient:        api.NewUserClient(b.StoreV2, auth),
		RBACClient:        api.NewRBACClient(b.StoreV2, auth),
		VersionController: actions.NewVersionController(clusterVersion),
		MetricGatherer:    prometheus.DefaultGatherer,
		GenericClient:     &api.GenericClient{Store: b.StoreV2, Auth: auth},
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
		Store:               b.StoreV2,
		EventStore:          b.EventStore,
		QueueGetter:         queueGetter,
		TLS:                 config.TLS,
		Cluster:             client.Cluster,
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
			Store:      b.StoreV2,
			EventStore: b.EventStore,
			RingPool:   b.RingPool,
			Client:     client,
			Bus:        bus,
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
		Store:         b.StoreV2,
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
