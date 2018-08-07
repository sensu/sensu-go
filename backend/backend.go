package backend

import (
	"crypto/tls"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/agentd"
	"github.com/sensu/sensu-go/backend/apid"
	"github.com/sensu/sensu-go/backend/daemon"
	"github.com/sensu/sensu-go/backend/dashboardd"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/eventd"
	"github.com/sensu/sensu-go/backend/keepalived"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/migration"
	"github.com/sensu/sensu-go/backend/monitor"
	"github.com/sensu/sensu-go/backend/pipelined"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/backend/ring"
	"github.com/sensu/sensu-go/backend/schedulerd"
	"github.com/sensu/sensu-go/backend/seeds"
	"github.com/sensu/sensu-go/backend/store"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/rpc"
	"github.com/sensu/sensu-go/types"
)

// Backend represents the backend server, which is used to hold the datastore
// and coordinating the daemons
type Backend struct {
	Daemons []daemon.Daemon
	Etcd    *etcd.Etcd
	Store   store.Store

	done         chan struct{}
	shutdownChan chan struct{}
}

// Initialize instantiates a Backend struct with the provided config, by
// configuring etcd and establishing a list of daemons, which constitute our
// backend. The daemons will later be started according to their position in the
// b.Daemons list, and stopped in reverse order
func Initialize(config *Config) (*Backend, error) {
	// Initialize a Backend struct
	b := &Backend{}

	b.done = make(chan struct{})
	b.shutdownChan = make(chan struct{})

	// Intialize the TLS configuration
	var (
		tlsConfig *tls.Config
		err       error
	)
	if config.TLS != nil {
		tlsConfig, err = config.TLS.ToTLSConfig()
		if err != nil {
			return nil, err
		}
	}

	// Initialize and start etcd, because we'll need to provide an etcd client to
	// the Wizard bus, which requires etcd to be started.
	cfg := etcd.NewConfig()
	cfg.DataDir = config.StateDir
	cfg.ListenClientURL = config.EtcdListenClientURL
	cfg.ListenPeerURL = config.EtcdListenPeerURL
	cfg.InitialCluster = config.EtcdInitialCluster
	cfg.InitialClusterState = config.EtcdInitialClusterState
	cfg.InitialAdvertisePeerURL = config.EtcdInitialAdvertisePeerURL
	cfg.Name = config.EtcdName

	if config.TLS != nil {
		cfg.TLSConfig = &etcd.TLSConfig{
			Info: etcd.TLSInfo{
				CertFile:      config.TLS.CertFile,
				KeyFile:       config.TLS.KeyFile,
				TrustedCAFile: config.TLS.TrustedCAFile,
			},
			TLS: tlsConfig,
		}
	}

	// Start etcd
	e, err := etcd.NewEtcd(cfg)
	if err != nil {
		return nil, errors.New("error starting etcd: " + err.Error())
	}

	// Create an etcd client for our daemons
	client, err := e.NewClient()
	if err != nil {
		return nil, errors.New("error initializing an etcd client: " + err.Error())
	}

	// Initialize the store, which lives on top of etcd
	store := etcdstore.NewStore(client, e.Name())
	if err = seeds.SeedInitialData(store); err != nil {
		return nil, errors.New("error initializing the store: " + err.Error())
	}

	// Initialize an etcd getter
	queueGetter := queue.EtcdGetter{Client: client, BackendID: e.BackendID()}

	// Initialize the bus
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{
		RingGetter: ring.EtcdGetter{Client: client, BackendID: e.BackendID()},
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", bus.Name(), err.Error())
	}
	b.Daemons = append(b.Daemons, bus)

	// Initialize pipelined
	pipeline, err := pipelined.New(pipelined.Config{
		Store: store,
		Bus:   bus,
		ExtensionExecutorGetter: rpc.NewGRPCExtensionExecutor,
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", pipeline.Name(), err.Error())
	}
	b.Daemons = append(b.Daemons, pipeline)

	// Initialize eventd
	event, err := eventd.New(eventd.Config{
		Store:          store,
		Bus:            bus,
		MonitorFactory: monitor.EtcdFactory(client),
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", event.Name(), err.Error())
	}
	b.Daemons = append(b.Daemons, event)

	// Initialize schedulerd
	scheduler, err := schedulerd.New(schedulerd.Config{
		Store:       store,
		Bus:         bus,
		QueueGetter: queueGetter,
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", scheduler.Name(), err.Error())
	}
	b.Daemons = append(b.Daemons, scheduler)

	// Initialize agentd
	agent, err := agentd.New(agentd.Config{
		Host:  config.AgentHost,
		Port:  config.AgentPort,
		Bus:   bus,
		Store: store,
		TLS:   config.TLS,
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", agent.Name(), err.Error())
	}
	b.Daemons = append(b.Daemons, agent)

	// Initialize keepalived
	keepalive, err := keepalived.New(keepalived.Config{
		DeregistrationHandler: config.DeregistrationHandler,
		Bus:            bus,
		Store:          store,
		MonitorFactory: monitor.EtcdFactory(client),
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", keepalive.Name(), err.Error())
	}
	b.Daemons = append(b.Daemons, keepalive)

	// Initialize apid
	api, err := apid.New(apid.Config{
		Host:          config.APIHost,
		Port:          config.APIPort,
		Bus:           bus,
		Store:         store,
		QueueGetter:   queueGetter,
		TLS:           config.TLS,
		BackendStatus: b.Status,
		Cluster:       clientv3.NewCluster(client),
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", api.Name(), err.Error())
	}
	b.Daemons = append(b.Daemons, api)

	// Initialize dashboardd
	dashboard, err := dashboardd.New(dashboardd.Config{
		APIPort: config.APIPort,
		Host:    config.DashboardHost,
		Port:    config.DashboardPort,
		TLS:     config.TLS,
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing %s: %s", dashboard.Name(), err.Error())
	}
	b.Daemons = append(b.Daemons, dashboard)

	// Add etcd and store to our backend, since it's needed across the methods
	b.Etcd = e
	b.Store = store

	return b, nil
}

// Run starts all of the Backend server's daemons
func (b *Backend) Run() error {
	eg := errGroup{
		out: make(chan error),
	}
	sg := stopGroup{}

	// Loop across the daemons in order to start them, then add them to our groups
	for _, d := range b.Daemons {
		if err := d.Start(); err != nil {
			return fmt.Errorf("error starting %s: %s", d.Name(), err.Error())
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

	// Add etcd to our errGroup, since it's not included in the daemon list
	eg.errors = append(eg.errors, b.Etcd)
	eg.Go()

	select {
	case err := <-eg.Err():
		logger.Error(err.Error())
	case <-b.shutdownChan:
		logger.Info("backend shutting down")
	}

	var derr error
	logger.Info("shutting down etcd")
	defer func() {
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

	if err := sg.Stop(); err != nil {
		if derr == nil {
			derr = err
		}
	}

	// we allow inErrChan to leak to avoid panics from other
	// goroutines writing errors to either after shutdown has been initiated.
	close(b.done)

	return derr
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

// Migration performs the migration of data inside the store
func (b *Backend) Migration() error {
	logger.Infof("starting migration on the store with URL '%s'", b.Etcd.LoopbackURL())
	migration.Run(b.Etcd.LoopbackURL())
	return nil
}

// Status returns a map of component name to boolean healthy indicator.
func (b *Backend) Status() types.StatusMap {
	sm := map[string]bool{
		"store": b.Etcd.Healthy(),
	}

	for _, d := range b.Daemons {
		sm[d.Name()] = d.Status() == nil
	}

	return sm
}

// Stop the Backend cleanly.
func (b *Backend) Stop() {
	close(b.shutdownChan)
	<-b.done
}
