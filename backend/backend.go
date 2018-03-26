package backend

import (
	"crypto/tls"
	"fmt"
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
	"github.com/sensu/sensu-go/types"
)

const (
	// DefaultEtcdName is the default etcd member node name (single-node cluster only)
	DefaultEtcdName = "default"

	// DefaultEtcdClientURL is the default URL to listen for Etcd clients
	DefaultEtcdClientURL = "http://127.0.0.1:2379"

	// DefaultEtcdPeerURL is the default URL to listen for Etcd peers (single-node cluster only)
	DefaultEtcdPeerURL = "http://127.0.0.1:2380"
)

// Config specifies a Backend configuration.
type Config struct {
	// Backend Configuration
	StateDir string

	// Agentd Configuration
	AgentHost string
	AgentPort int

	// Apid Configuration
	APIHost string
	APIPort int

	// Dashboardd Configuration
	DashboardHost string
	DashboardPort int

	// Pipelined Configuration
	DeregistrationHandler string

	// Etcd configuration
	EtcdInitialAdvertisePeerURL string
	EtcdInitialClusterToken     string
	EtcdInitialClusterState     string
	EtcdInitialCluster          string
	EtcdListenClientURL         string
	EtcdListenPeerURL           string
	EtcdName                    string

	TLS *types.TLSOptions
}

// A Backend is a Sensu Backend server responsible for handling incoming
// HTTP requests and upgrading them
type Backend struct {
	Config *Config

	shutdownChan chan struct{}
	done         chan struct{}
	messageBus   messaging.MessageBus
	apid         daemon.Daemon
	agentd       daemon.Daemon
	schedulerd   daemon.Daemon
	etcd         *etcd.Etcd

	dashboardd daemon.Daemon
	eventd     daemon.Daemon
	pipelined  daemon.Daemon
	keepalived daemon.Daemon
}

// NewBackend will, given a Config, create an initialized Backend and return a
// pointer to it.
func NewBackend(config *Config) (*Backend, error) {
	// In other places we have a NewConfig() method, but I think that doing it
	// this way is more safe, because it doesn't require "trust" in callers.
	if config.EtcdListenClientURL == "" {
		config.EtcdListenClientURL = DefaultEtcdClientURL
	}

	if config.EtcdListenPeerURL == "" {
		config.EtcdListenPeerURL = DefaultEtcdPeerURL
	}

	if config.EtcdInitialCluster == "" {
		config.EtcdInitialCluster = fmt.Sprintf("%s=%s", DefaultEtcdName, DefaultEtcdPeerURL)
	}

	if config.EtcdInitialClusterState == "" {
		config.EtcdInitialClusterState = etcd.ClusterStateNew
	}

	if config.EtcdInitialAdvertisePeerURL == "" {
		config.EtcdInitialAdvertisePeerURL = DefaultEtcdPeerURL
	}

	if config.EtcdName == "" {
		config.EtcdName = DefaultEtcdName
	}

	if config.APIPort == 0 {
		config.APIPort = 8080
	}

	if config.AgentPort == 0 {
		config.AgentPort = 8081
	}

	// Check for TLS config and load certs if present
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

	b := &Backend{
		Config: config,

		done:         make(chan struct{}),
		shutdownChan: make(chan struct{}),
	}

	// we go ahead and setup and start etcd here, because we'll have to pass
	// a store along to the API.
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
		return nil, fmt.Errorf("error starting etcd: %s", err.Error())
	}
	b.etcd = e

	client, err := e.NewClient()
	if err != nil {
		return nil, err
	}

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{
		RingGetter: ring.EtcdGetter{Client: client},
	})
	if err != nil {
		return nil, err
	}
	b.messageBus = bus

	return b, nil
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

// Run starts all of the Backend server's event loops and sets up the HTTP
// server.
func (b *Backend) Run() (derr error) {
	if err := b.messageBus.Start(); err != nil {
		return err
	}

	// Right now, instantiating a new Etcd will start etcd. If we change that
	// s.t. Etcd has its own Start() method, conforming to Daemon, then we will
	// want to make sure that we aren't calling NewClient before starting it,
	// I think. That might return a connection error.
	client, err := b.etcd.NewClient()
	if err != nil {
		return err
	}
	etcdName := b.etcd.Name()

	// Seed initial data
	store := etcdstore.NewStore(client, etcdName)
	if err := seeds.SeedInitialData(store); err != nil {
		return err
	}

	bus := b.messageBus
	tlsOpts := b.Config.TLS
	queueGetter := queue.EtcdGetter{Client: client}

	b.schedulerd, err = schedulerd.New(schedulerd.Config{
		Store:       store,
		Bus:         bus,
		QueueGetter: queueGetter,
	})
	if err != nil {
		return fmt.Errorf("error creating schedulerd: %s", err)
	}
	if err := b.schedulerd.Start(); err != nil {
		return fmt.Errorf("error starting schedulerd: %s", err)
	}

	b.pipelined, err = pipelined.New(pipelined.Config{
		Store: store,
		Bus:   bus,
	})
	if err != nil {
		return fmt.Errorf("error creating pipelined: %s", err)
	}

	if err := b.pipelined.Start(); err != nil {
		return fmt.Errorf("error starting pipelined: %s", err)
	}

	// TLS config gets passed down here
	b.apid, err = apid.New(apid.Config{
		Host:          b.Config.APIHost,
		Port:          b.Config.APIPort,
		Bus:           bus,
		Store:         store,
		QueueGetter:   queueGetter,
		TLS:           tlsOpts,
		BackendStatus: b.Status,
	})
	if err != nil {
		return fmt.Errorf("error creating apid: %s", err)
	}

	if err := b.apid.Start(); err != nil {
		return fmt.Errorf("error starting apid: %s", err)
	}

	b.agentd, err = agentd.New(agentd.Config{
		Host:  b.Config.AgentHost,
		Port:  b.Config.AgentPort,
		Bus:   bus,
		Store: store,
		TLS:   tlsOpts,
	})
	if err != nil {
		return fmt.Errorf("error creating agentd: %s", err)
	}

	if err := b.agentd.Start(); err != nil {
		return fmt.Errorf("error starting agentd: %s", err)
	}

	dashCfg := dashboardd.Config{
		Host: b.Config.DashboardHost,
		Port: b.Config.DashboardPort,
		TLS:  b.Config.TLS,
	}
	b.dashboardd, err = dashboardd.New(dashCfg)
	if err != nil {
		return fmt.Errorf("error creating dashboardd: %s", err)
	}

	if err := b.dashboardd.Start(); err != nil {
		return fmt.Errorf("error starting dashboardd: %s", err)
	}

	b.eventd, err = eventd.New(eventd.Config{
		Store: store,
		Bus:   bus,
	})
	if err != nil {
		return fmt.Errorf("error creating eventd: %s", err)
	}

	if err := b.eventd.Start(); err != nil {
		return fmt.Errorf("error starting eventd: %s", err)
	}

	b.keepalived, err = keepalived.New(keepalived.Config{
		DeregistrationHandler: b.Config.DeregistrationHandler,
		Bus:   bus,
		Store: store,
	})
	if err != nil {
		return fmt.Errorf("error creating keepalived: %s", err)
	}

	if err := b.keepalived.Start(); err != nil {
		return fmt.Errorf("error starting keepalived: %s", err)
	}

	eg := errGroup{
		out: make(chan error),
		errors: []errorer{
			b.apid,
			b.agentd,
			b.schedulerd,
			b.etcd,
			b.messageBus,
			b.pipelined,
			b.dashboardd,
			b.eventd,
			b.keepalived,
		},
	}
	eg.Go()

	select {
	case err := <-eg.Err():
		logger.Error(err.Error())
	case <-b.shutdownChan:
		logger.Info("backend shutting down")
	}

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

	sg := stopGroup{
		// stop allowing API connections
		{Name: "apid", stopper: b.apid},
		// stop allowing dashboard connections
		{Name: "dashboardd", stopper: b.dashboardd},
		// disconnect all agents and don't allow any more to connect.
		{Name: "agentd", stopper: b.agentd},
		// stop scheduling checks.
		{Name: "schedulerd", stopper: b.schedulerd},
		// Shutting down eventd will cause it to drain events to the bus
		{Name: "eventd", stopper: b.eventd},
		// Once events have been drained from eventd, pipelined can finish
		// processing events.
		{Name: "pipelined", stopper: b.pipelined},
		// finally shutdown the message bus once all other components have stopped
		// using it.
		{Name: "message bus", stopper: b.messageBus},
	}

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

// Migration performs the migration of data inside the store
func (b *Backend) Migration() error {
	logger.Infof("starting migration on the store with URL '%s'", b.etcd.LoopbackURL())
	migration.Run(b.etcd.LoopbackURL())
	return nil
}

// Status returns a map of component name to boolean healthy indicator.
func (b *Backend) Status() types.StatusMap {
	sm := map[string]bool{
		"store":       b.etcd.Healthy(),
		"message_bus": b.messageBus.Status() == nil,
		"schedulerd":  b.schedulerd.Status() == nil,
		"pipelined":   b.pipelined.Status() == nil,
		"eventd":      b.eventd.Status() == nil,
		"agentd":      b.agentd.Status() == nil,
		"apid":        b.apid.Status() == nil,
	}

	return sm
}

// Stop the Backend cleanly.
func (b *Backend) Stop() {
	close(b.shutdownChan)
	<-b.done
}
