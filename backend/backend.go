package backend

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/prometheus/client_golang/prometheus"
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
	"github.com/sensu/sensu-go/backend/dashboardd"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/eventd"
	"github.com/sensu/sensu-go/backend/keepalived"
	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/pipelined"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/schedulerd"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/backend/tessend"
	"github.com/sensu/sensu-go/rpc"
	"github.com/sensu/sensu-go/system"
	"github.com/sensu/sensu-go/util/retry"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type ErrStartup struct {
	Err  error
	Name string
}

func (e ErrStartup) Error() string {
	return fmt.Sprintf("error starting %s: %s", e.Name, e.Err)
}

// Backend represents the backend server, which is used to hold the datastore
// and coordinating the daemons
type Backend struct {
	Client                 *clientv3.Client
	Daemons                []daemon.Daemon
	Etcd                   *etcd.Etcd
	Store                  store.Store
	EventStore             EventStoreUpdater
	GraphQLService         *graphql.Service
	SecretsProviderManager *secrets.ProviderManager
	HealthRouter           *routers.HealthRouter
	EtcdClientTLSConfig    *tls.Config

	ctx       context.Context
	runCtx    context.Context
	runCancel context.CancelFunc
	cfg       *Config
}

// EventStoreUpdater offers a way to update an event store to a different
// implementation in-place.
type EventStoreUpdater interface {
	UpdateEventStore(to store.EventStore)
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

		// Don't start up an embedded etcd, return a client that connects to an
		// external etcd instead.
		client, err := clientv3.New(clientv3.Config{
			Endpoints:   clientURLs,
			DialTimeout: 5 * time.Second,
			TLS:         tlsConfig,
			DialOptions: []grpc.DialOption{
				grpc.WithBlock(),
			},
		})
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

	// Start etcd
	e, err := etcd.NewEtcd(cfg)
	if err != nil {
		return nil, fmt.Errorf("error starting etcd: %s", err)
	}

	backend.Etcd = e

	// Create an etcd client
	var client *clientv3.Client
	if config.EtcdUseEmbeddedClient {
		client = e.NewEmbeddedClient()
	} else {
		cl, err := e.NewClient()
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
	b := &Backend{cfg: config}

	b.ctx = ctx
	b.runCtx, b.runCancel = context.WithCancel(b.ctx)

	b.Client, err = newClient(b.runCtx, config, b)
	if err != nil {
		return nil, err
	}

	// Create the store, which lives on top of etcd
	stor := etcdstore.NewStore(b.Client, config.EtcdName)
	b.Store = stor

	if _, err := stor.GetClusterID(b.runCtx); err != nil {
		return nil, err
	}

	// Initialize the JWT secret. This method is idempotent and needs to be ran
	// at every startup so the JWT signatures remain valid
	if err := jwt.InitSecret(b.Store); err != nil {
		return nil, err
	}

	eventStoreProxy := store.NewEventStoreProxy(stor)
	b.EventStore = eventStoreProxy

	logger.Debug("Registering backend...")

	backendID := etcd.NewBackendIDGetter(b.runCtx, b.Client)
	logger.Debug("Done registering backend.")
	b.Daemons = append(b.Daemons, backendID)

	// Initialize an etcd getter
	queueGetter := queue.EtcdGetter{Client: b.Client, BackendIDGetter: backendID}

	// Initialize the bus
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", bus.Name(), err)
	}
	b.Daemons = append(b.Daemons, bus)

	// Initialize asset manager
	backendEntity := b.getBackendEntity(config)
	logger.WithField("entity", backendEntity).Info("backend entity information")
	assetManager := asset.NewManager(config.CacheDir, backendEntity, &sync.WaitGroup{})
	assetGetter, err := assetManager.StartAssetManager(b.runCtx)
	if err != nil {
		return nil, fmt.Errorf("error initializing asset manager: %s", err)
	}

	// Initialize the secrets provider manager
	b.SecretsProviderManager = secrets.NewProviderManager()

	// Initialize pipelined
	pipeline, err := pipelined.New(pipelined.Config{
		Store: stor,
		Bus:   bus,
		ExtensionExecutorGetter: rpc.NewGRPCExtensionExecutor,
		AssetGetter:             assetGetter,
		BufferSize:              viper.GetInt(FlagPipelinedBufferSize),
		WorkerCount:             viper.GetInt(FlagPipelinedWorkers),
		StoreTimeout:            2 * time.Minute,
		SecretsProviderManager:  b.SecretsProviderManager,
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", pipeline.Name(), err)
	}
	b.Daemons = append(b.Daemons, pipeline)

	// Initialize eventd
	event, err := eventd.New(
		b.runCtx,
		eventd.Config{
			Store:           stor,
			EventStore:      eventStoreProxy,
			Bus:             bus,
			LivenessFactory: liveness.EtcdFactory(b.runCtx, b.Client),
			Client:          b.Client,
			BufferSize:      viper.GetInt(FlagEventdBufferSize),
			WorkerCount:     viper.GetInt(FlagEventdWorkers),
			StoreTimeout:    2 * time.Minute,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", event.Name(), err)
	}
	b.Daemons = append(b.Daemons, event)

	ringPool := ringv2.NewPool(b.Client)

	// Initialize schedulerd
	scheduler, err := schedulerd.New(
		b.runCtx,
		schedulerd.Config{
			Store:                  stor,
			Bus:                    bus,
			QueueGetter:            queueGetter,
			RingPool:               ringPool,
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

	// Initialize agentd
	agent, err := agentd.New(agentd.Config{
		Host:         config.AgentHost,
		Port:         config.AgentPort,
		Bus:          bus,
		Store:        stor,
		TLS:          config.AgentTLSOptions,
		RingPool:     ringPool,
		WriteTimeout: config.AgentWriteTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", agent.Name(), err)
	}
	b.Daemons = append(b.Daemons, agent)

	// Initialize keepalived
	keepalive, err := keepalived.New(keepalived.Config{
		DeregistrationHandler: config.DeregistrationHandler,
		Bus:             bus,
		Store:           stor,
		EventStore:      stor,
		LivenessFactory: liveness.EtcdFactory(b.runCtx, b.Client),
		RingPool:        ringPool,
		BufferSize:      viper.GetInt(FlagKeepalivedBufferSize),
		WorkerCount:     viper.GetInt(FlagKeepalivedWorkers),
		StoreTimeout:    2 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", keepalive.Name(), err)
	}
	b.Daemons = append(b.Daemons, keepalive)

	// Prepare the etcd client TLS config
	etcdClientTLSInfo := (transport.TLSInfo)(config.EtcdClientTLSInfo)
	etcdClientTLSConfig, err := etcdClientTLSInfo.ClientConfig()
	if err != nil {
		return nil, err
	}
	b.EtcdClientTLSConfig = etcdClientTLSConfig

	// Prepare the authentication providers
	authenticator := &authentication.Authenticator{}
	basic := &basic.Provider{
		ObjectMeta: corev2.ObjectMeta{Name: basic.Type},
		Store:      stor,
	}
	authenticator.AddProvider(basic)

	var clusterVersion string
	// only retrieve the cluster version if etcd is embedded
	if !config.NoEmbedEtcd {
		clusterVersion = b.Etcd.GetClusterVersion()
	}

	// Load the JWT key pair
	if err := jwt.LoadKeyPair(viper.GetString(FlagJWTPrivateKeyFile), viper.GetString(FlagJWTPublicKeyFile)); err != nil {
		logger.WithError(err).Error("could not load the key pair for the JWT signature")
	}

	// Initialize the health router
	b.HealthRouter = routers.NewHealthRouter(actions.NewHealthController(stor, b.Client.Cluster, b.EtcdClientTLSConfig))

	// Initialize GraphQL service
	auth := &rbac.Authorizer{Store: stor}
	b.GraphQLService, err = graphql.NewService(graphql.ServiceConfig{
		AssetClient:       api.NewAssetClient(stor, auth),
		CheckClient:       api.NewCheckClient(stor, actions.NewCheckController(stor, queueGetter), auth),
		EntityClient:      api.NewEntityClient(stor, eventStoreProxy, auth),
		EventClient:       api.NewEventClient(eventStoreProxy, auth, bus),
		EventFilterClient: api.NewEventFilterClient(stor, auth),
		HandlerClient:     api.NewHandlerClient(stor, auth),
		HealthController:  actions.NewHealthController(stor, b.Client.Cluster, etcdClientTLSConfig),
		MutatorClient:     api.NewMutatorClient(stor, auth),
		SilencedClient:    api.NewSilencedClient(stor, auth),
		NamespaceClient:   api.NewNamespaceClient(stor, auth),
		HookClient:        api.NewHookConfigClient(stor, auth),
		UserClient:        api.NewUserClient(stor, auth),
		RBACClient:        api.NewRBACClient(stor, auth),
		VersionController: actions.NewVersionController(clusterVersion),
		MetricGatherer:    prometheus.DefaultGatherer,
		GenericClient:     &api.GenericClient{Store: stor, Auth: auth},
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing graphql.Service: %s", err)
	}

	// Initialize apid
	apidConfig := apid.Config{
		ListenAddress:       config.APIListenAddress,
		URL:                 config.APIURL,
		Bus:                 bus,
		Store:               stor,
		EventStore:          eventStoreProxy,
		QueueGetter:         queueGetter,
		TLS:                 config.TLS,
		Cluster:             b.Client.Cluster,
		EtcdClientTLSConfig: etcdClientTLSConfig,
		Authenticator:       authenticator,
		ClusterVersion:      clusterVersion,
		GraphQLService:      b.GraphQLService,
		HealthRouter:        b.HealthRouter,
	}
	api, err := apid.New(apidConfig)
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", api.Name(), err)
	}
	b.Daemons = append(b.Daemons, api)

	// Initialize tessend
	tessen, err := tessend.New(
		b.runCtx,
		tessend.Config{
			Store:      stor,
			EventStore: eventStoreProxy,
			RingPool:   ringPool,
			Client:     b.Client,
			Bus:        bus,
		})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", tessen.Name(), err)
	}
	b.Daemons = append(b.Daemons, tessen)

	// Initialize dashboardd TLS config
	var dashboardTLSConfig *corev2.TLSOptions

	// Always use dashboard tls options when they are specified
	if config.DashboardTLSCertFile != "" && config.DashboardTLSKeyFile != "" {
		dashboardTLSConfig = &corev2.TLSOptions{
			CertFile: config.DashboardTLSCertFile,
			KeyFile:  config.DashboardTLSKeyFile,
		}
	} else if config.TLS != nil {
		// use apid tls config if no dashboard tls options are specified
		dashboardTLSConfig = &corev2.TLSOptions{
			CertFile: config.TLS.GetCertFile(),
			KeyFile:  config.TLS.GetKeyFile(),
		}
	}
	dashboard, err := dashboardd.New(dashboardd.Config{
		APIDConfig: apidConfig,
		Host:       config.DashboardHost,
		Port:       config.DashboardPort,
		TLS:        dashboardTLSConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", dashboard.Name(), err)
	}
	b.Daemons = append(b.Daemons, dashboard)

	return b, nil
}

func (b *Backend) runOnce() error {
	var derr error

	if b.Etcd != nil {
		defer func() {
			logger.Info("shutting down etcd")
			if err := recover(); err != nil {
				trace := string(debug.Stack())
				logger.WithField("panic", trace).WithError(err.(error)).
					Error("recovering from panic due to error, shutting down etcd")
			}
			err := b.Etcd.Shutdown()
			if derr == nil {
				derr = err
			}
		}()
	}

	eg := errGroup{
		out: make(chan error),
	}
	sg := stopGroup{}

	// Loop across the daemons in order to start them, then add them to our groups
	for _, d := range b.Daemons {
		if err := d.Start(); err != nil {
			return ErrStartup{Err: err, Name: d.Name()}
		}

		// Add the daemon to our errGroup
		eg.errors = append(eg.errors, d)

		// Add the daemon to our stopGroup
		sg = append(sg, daemonStopper{
			Name:    d.Name(),
			stopper: d,
		})
	}

	// Reverse the order of our stopGroup so daemons are stopped in the proper
	// order (last one started is first one stopped)
	for i := len(sg)/2 - 1; i >= 0; i-- {
		opp := len(sg) - 1 - i
		sg[i], sg[opp] = sg[opp], sg[i]
	}

	if b.Etcd != nil {
		// Add etcd to our errGroup, since it's not included in the daemon list
		eg.errors = append(eg.errors, b.Etcd)
	}
	eg.Go()

	select {
	case err := <-eg.Err():
		logger.WithError(err).Error("backend stopped working and is restarting")
	case <-b.runCtx.Done():
		logger.Info("backend shutting down")
	}
	if err := sg.Stop(); err != nil {
		if derr == nil {
			derr = err
		}
	}

	if derr == nil {
		derr = b.runCtx.Err()
	}

	_ = b.Client.Close()

	return derr
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

		// Yes, two levels of retry... this could improve. Unfortunately Intialize()
		// is called elsewhere.
		err = backoff.Retry(func(int) (bool, error) {
			backend, err := initialize(b.ctx, b.cfg)
			if err != nil && err != context.Canceled {
				logger.Error(err)
				return false, nil
			} else if err == context.Canceled {
				return true, err
			}
			// Replace b with a new backend - this is done to ensure that there is
			// no side effects from the execution of b that have carried over
			*b = *backend
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
}

type daemonStopper struct {
	stopper
	Name string
}

type stopGroup []daemonStopper

func (s stopGroup) Stop() (err error) {
	for _, stopper := range s {
		logger.Info("shutting down ", stopper.Name)
		e := stopper.Stop()
		if err == nil {
			err = e
		}
	}
	return err
}

type errorer interface {
	Err() <-chan error
}

type errGroup struct {
	out    chan error
	errors []errorer
}

func (e errGroup) Go() {
	for _, err := range e.errors {
		err := err
		go func() {
			e.out <- <-err.Err()
		}()
	}
}

func (e errGroup) Err() <-chan error {
	return e.out
}

// Stop the Backend cleanly.
func (b *Backend) Stop() {
	b.runCancel()
}

func (b *Backend) getBackendEntity(config *Config) *corev2.Entity {
	entity := &corev2.Entity{
		EntityClass: corev2.EntityBackendClass,
		System:      getSystemInfo(),
		ObjectMeta:  corev2.NewObjectMeta(getDefaultBackendID(), ""),
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
