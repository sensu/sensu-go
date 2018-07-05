package backend

import (
	"crypto/tls"
	"errors"
	"runtime/debug"

	"github.com/sensu/sensu-go/backend/agentd"
	"github.com/sensu/sensu-go/backend/apid"
	"github.com/sensu/sensu-go/backend/daemon"
	"github.com/sensu/sensu-go/backend/dashboardd"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/eventd"
	"github.com/sensu/sensu-go/backend/keepalived"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/migration"
	"github.com/sensu/sensu-go/backend/pipelined"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/backend/ring"
	"github.com/sensu/sensu-go/backend/schedulerd"
	"github.com/sensu/sensu-go/backend/seeds"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/rpc"
	"github.com/sensu/sensu-go/types"
)

// Backend is a Sensu Backend server responsible for handling incoming
// HTTP requests and upgrading them
type Backend struct {
	daemons []daemon.Daemon
	etcd    *etcd.Etcd

	done         chan struct{}
	shutdownChan chan struct{}
}

// Initialize ...
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

	e, err := etcd.NewEtcd(cfg)
	if err != nil {
		return nil, errors.New("error starting etcd: " + err.Error())
	}

	// Create an etcd client for our daemons
	client, err := e.NewClient()
	if err != nil {
		return nil, errors.New("error initializing an etcd client: " + err.Error())
	}

	// Initialize the bus
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{
		RingGetter: ring.EtcdGetter{Client: client},
	})
	if err != nil {
		return nil, errors.New("error initializing the bus: " + err.Error())
	}
	b.daemons = append(b.daemons, bus)

	// Initialize the store, which lives on top of etcd
	store := etcdstore.NewStore(client, e.Name())
	if err := seeds.SeedInitialData(store); err != nil {
		return nil, errors.New("error initializing the store: " + err.Error())
	}

	// Initialize an etcd getter
	queueGetter := queue.EtcdGetter{Client: client}

	// Initialize schedulerd
	scheduler, err := schedulerd.New(schedulerd.Config{
		Store:       store,
		Bus:         bus,
		QueueGetter: queueGetter,
	})
	if err != nil {
		return nil, errors.New("error initializing schedulerd: " + err.Error())
	}
	b.daemons = append(b.daemons, scheduler)

	// Initialize pipelined
	pipeline, err := pipelined.New(pipelined.Config{
		Store: store,
		Bus:   bus,
		ExtensionExecutorGetter: rpc.NewGRPCExtensionExecutor,
	})
	if err != nil {
		return nil, errors.New("error initializing pipelined: " + err.Error())
	}
	b.daemons = append(b.daemons, pipeline)

	// Initialize apid
	api, err := apid.New(apid.Config{
		Host:          config.APIHost,
		Port:          config.APIPort,
		Bus:           bus,
		Store:         store,
		QueueGetter:   queueGetter,
		TLS:           config.TLS,
		BackendStatus: b.Status,
	})
	if err != nil {
		return nil, errors.New("error initializing apid: " + err.Error())
	}
	b.daemons = append(b.daemons, api)

	// Initialize agentd
	agent, err := agentd.New(agentd.Config{
		Host:  config.AgentHost,
		Port:  config.AgentPort,
		Bus:   bus,
		Store: store,
		TLS:   config.TLS,
	})
	if err != nil {
		return nil, errors.New("error initializing agentd: " + err.Error())
	}
	b.daemons = append(b.daemons, agent)

	// Initialize dashboardd
	dashboard, err := dashboardd.New(dashboardd.Config{
		APIPort: config.APIPort,
		Host:    config.DashboardHost,
		Port:    config.DashboardPort,
		TLS:     config.TLS,
	})
	if err != nil {
		return nil, errors.New("error initializing dashboardd: " + err.Error())
	}
	b.daemons = append(b.daemons, dashboard)

	// Initialize eventd
	event, err := eventd.New(eventd.Config{
		Store: store,
		Bus:   bus,
	})
	if err != nil {
		return nil, errors.New("error initializing eventd: " + err.Error())
	}
	b.daemons = append(b.daemons, event)

	// Initialize keepalived
	keepalive, err := keepalived.New(keepalived.Config{
		DeregistrationHandler: config.DeregistrationHandler,
		Bus:   bus,
		Store: store,
	})
	if err != nil {
		return nil, errors.New("error initializing keepalived: " + err.Error())
	}
	b.daemons = append(b.daemons, keepalive)

	// Add etcd to our backend, since it's needed almost everywhere
	b.etcd = e

	return b, nil
}

// Run ...
func (b *Backend) Run() error {
	eg := errGroup{
		out: make(chan error),
	}
	sg := stopGroup{}

	for _, d := range b.daemons {
		if err := d.Start(); err != nil {
			return errors.New("error starting daemon: " + err.Error())
		}

		// Add the daemon to our errGroup
		eg.errors = append(eg.errors, d)

		// Add the daemon to our stopGroup
		sg = append(sg, daemonStopper{
			Name:    d.Name(),
			stopper: d,
		})
	}

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
			logger.WithField("panic", trace).WithError(err.(error)).Error("recovering from panic due to error, shutting down etcd")
		}
		err := b.etcd.Shutdown()
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
	logger.Infof("starting migration on the store with URL '%s'", b.etcd.LoopbackURL())
	migration.Run(b.etcd.LoopbackURL())
	return nil
}

// Status returns a map of component name to boolean healthy indicator.
func (b *Backend) Status() types.StatusMap {
	sm := map[string]bool{
		"store": b.etcd.Healthy(),
	}

	for _, d := range b.daemons {
		sm[d.Name()] = d.Status() == nil
	}

	return sm
}

// Stop the Backend cleanly.
func (b *Backend) Stop() {
	close(b.shutdownChan)
	<-b.done
}
