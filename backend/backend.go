package backend

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sensu/sensu-go/backend/resource"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"

	corev2 "github.com/sensu/core/v2"
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
	"github.com/sensu/sensu-go/backend/eventd"
	"github.com/sensu/sensu-go/backend/keepalived"
	"github.com/sensu/sensu-go/backend/licensing"
	"github.com/sensu/sensu-go/backend/logging"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/pipeline"
	"github.com/sensu/sensu-go/backend/pipeline/filter"
	"github.com/sensu/sensu-go/backend/pipeline/handler"
	"github.com/sensu/sensu-go/backend/pipeline/mutator"
	"github.com/sensu/sensu-go/backend/pipelined"
	"github.com/sensu/sensu-go/backend/schedulerd"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store/postgres"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
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
	"graphql_duration_seconds",
	"graphql_duration_seconds_sum",
	"graphql_duration_seconds_count",
}

// Backend represents the backend server, which is used to hold the datastore
// and coordinating the daemons
type Backend struct {
	Daemons                []daemon.Daemon
	Store                  storev2.Interface
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

// Initialize instantiates a Backend struct with the provided config, which creates
// a list of daemons. The daemons will be started according to their position in the
// b.Daemons list, and stopped in reverse order
func Initialize(ctx context.Context, pgdb postgres.DBI, config *Config) (*Backend, error) {
	var err error
	// Initialize a Backend struct
	b := &Backend{Cfg: config}

	b.Store = postgres.NewStore(postgres.StoreConfig{
		DB:             pgdb,
		WatchInterval:  time.Second,
		WatchTxnWindow: 5 * time.Second,
	})

	jwtClient := api.JWT{Store: b.Store}
	jwtSecret, err := jwtClient.GetSecret(ctx)
	if err != nil {
		return nil, err
	}
	// TODO: don't use global variables
	jwt.SetSecret(jwtSecret)

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
	br := resource.New(b.Store.GetNamespaceStore(), b.Store.GetEntityConfigStore(), b.Store.GetEntityStateStore(), bus)
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

	pgOPC := postgres.NewOPC(pgdb)

	// Initialize eventd
	event, err := eventd.New(
		ctx,
		eventd.Config{
			Store:               b.Store,
			Bus:                 bus,
			BufferSize:          viper.GetInt(FlagEventdBufferSize),
			WorkerCount:         viper.GetInt(FlagEventdWorkers),
			StoreTimeout:        2 * time.Minute,
			LogPath:             b.Cfg.EventLogFile,
			LogBufferSize:       b.Cfg.EventLogBufferSize,
			LogBufferWait:       b.Cfg.EventLogBufferWait,
			LogParallelEncoders: b.Cfg.EventLogParallelEncoders,
			OperatorConcierge:   pgOPC,
			OperatorMonitor:     pgOPC,
			OperatorQueryer:     pgOPC,
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
			Store:                  b.Store,
			Bus:                    bus,
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
	entityConfigWatcher := agentd.GetEntityConfigWatcher(ctx, b.Store)

	// Initialize keepalived
	keepalive, err := keepalived.New(keepalived.Config{
		DeregistrationHandler: config.DeregistrationHandler,
		Bus:                   bus,
		Store:                 b.Store,
		BufferSize:            viper.GetInt(FlagKeepalivedBufferSize),
		WorkerCount:           viper.GetInt(FlagKeepalivedWorkers),
		StoreTimeout:          2 * time.Minute,
		OperatorConcierge:     pgOPC,
		OperatorMonitor:       pgOPC,
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

	// Load the JWT key pair
	if err := jwt.LoadKeyPair(viper.GetString(FlagJWTPrivateKeyFile), viper.GetString(FlagJWTPublicKeyFile)); err != nil {
		logger.WithError(err).Error("could not load the key pair for the JWT signature")
	}

	// Initialize the health router
	b.HealthRouter = routers.NewHealthRouter(actions.HealthController{})

	// Initialize GraphQL service
	b.GraphQLService, err = graphql.NewService(graphql.ServiceConfig{
		AssetClient:       api.NewAssetClient(b.Store, auth),
		CheckClient:       api.NewCheckClient(b.Store, actions.NewCheckController(b.Store, nil), auth),
		EntityClient:      api.NewEntityClient(b.Store, auth),
		EventClient:       api.NewEventClient(b.Store.GetEventStore(), auth, bus),
		EventFilterClient: api.NewEventFilterClient(b.Store, auth),
		HandlerClient:     api.NewHandlerClient(b.Store, auth),
		HealthController:  actions.HealthController{},
		MutatorClient:     api.NewMutatorClient(b.Store, auth),
		SilencedClient:    api.NewSilencedClient(b.Store.GetSilencesStore(), auth),
		NamespaceClient:   api.NewNamespaceClient(b.Store, auth),
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
		ListenAddress:  config.APIListenAddress,
		RequestLimit:   config.APIRequestLimit,
		WriteTimeout:   config.APIWriteTimeout,
		URL:            config.APIURL,
		Bus:            bus,
		Store:          b.Store,
		TLS:            config.TLS,
		Authenticator:  authenticator,
		ClusterVersion: clusterVersion,
		GraphQLService: b.GraphQLService,
		HealthRouter:   b.HealthRouter,
	}
	newApi, err := apid.New(b.APIDConfig)
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", newApi.Name(), err)
	}
	b.Daemons = append(b.Daemons, newApi)

	// Initialize tessend
	// TODO(eric): add support for non-etcd tessend
	// tessen, err := tessend.New(
	// 	ctx,
	// 	tessend.Config{
	// 		Store:      b.ConfigStore,
	// 		EventStore: b.EventStore,
	// 		RingPool:   b.RingPool,
	// 		Client:     etcdConfigClient,
	// 		Bus:        bus,
	// 		ClusterID:  clusterID,
	// 	})
	// if err != nil {
	// 	return nil, fmt.Errorf("error initializing %s: %s", tessen.Name(), err)
	// }
	// b.Daemons = append(b.Daemons, tessen)

	// Initialize agentd
	agent, err := agentd.New(agentd.Config{
		Host:          config.AgentHost,
		Port:          config.AgentPort,
		Bus:           bus,
		Store:         b.Store,
		TLS:           config.AgentTLSOptions,
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

	// crash the stopgroup after a hard-coded timeout
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
