package backend

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/sensu/sensu-go/backend/agentd"
	"github.com/sensu/sensu-go/backend/apid"
	"github.com/sensu/sensu-go/backend/daemon"
	"github.com/sensu/sensu-go/backend/dashboardd"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/eventd"
	"github.com/sensu/sensu-go/backend/keepalived"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/monitor"
	"github.com/sensu/sensu-go/backend/pipelined"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/backend/ring"
	"github.com/sensu/sensu-go/backend/schedulerd"
	"github.com/sensu/sensu-go/backend/seeds"
	"github.com/sensu/sensu-go/backend/store"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/rpc"
)

// Backend represents the backend server, which is used to hold the datastore
// and coordinating the daemons
type Backend struct {
	Daemons []daemon.Daemon
	Etcd    *etcd.Etcd
	Store   store.Store

	done   chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

func newClient(config *Config, backend *Backend) (*clientv3.Client, error) {
	if config.NoEmbedEtcd {
		tlsInfo := (transport.TLSInfo)(config.EtcdClientTLSInfo)
		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return nil, err
		}

		// Don't start up an embedded etcd, return a client that connects to an
		// external etcd instead.
		return clientv3.New(clientv3.Config{
			Endpoints:   strings.Split(config.EtcdListenClientURL, ","),
			DialTimeout: 5 * time.Second,
			TLS:         tlsConfig,
		})
	}

	// Initialize and start etcd, because we'll need to provide an etcd client to
	// the Wizard bus, which requires etcd to be started.
	cfg := etcd.NewConfig()
	cfg.DataDir = config.StateDir
	cfg.ListenClientURLs = strings.Split(config.EtcdListenClientURL, ",")
	cfg.ListenPeerURL = config.EtcdListenPeerURL
	cfg.InitialCluster = config.EtcdInitialCluster
	cfg.InitialClusterState = config.EtcdInitialClusterState
	cfg.InitialAdvertisePeerURL = config.EtcdInitialAdvertisePeerURL
	cfg.Name = config.EtcdName

	// Etcd TLS config
	cfg.ClientTLSInfo = config.EtcdClientTLSInfo
	cfg.PeerTLSInfo = config.EtcdPeerTLSInfo

	// Start etcd
	e, err := etcd.NewEtcd(cfg)
	if err != nil {
		return nil, fmt.Errorf("error starting etcd: %s", err)
	}

	backend.Etcd = e

	// Create an etcd client
	return e.NewClient()
}

// Initialize instantiates a Backend struct with the provided config, by
// configuring etcd and establishing a list of daemons, which constitute our
// backend. The daemons will later be started according to their position in the
// b.Daemons list, and stopped in reverse order
func Initialize(config *Config) (*Backend, error) {
	// Initialize a Backend struct
	b := &Backend{}

	b.done = make(chan struct{})
	b.ctx, b.cancel = context.WithCancel(context.Background())

	client, err := newClient(config, b)
	if err != nil {
		return nil, err
	}

	// Initialize the store, which lives on top of etcd
	logger.Debug("Initializing store...")
	store := etcdstore.NewStore(client, config.EtcdName)
	if err = seeds.SeedInitialData(store); err != nil {
		return nil, errors.New("error initializing the store: " + err.Error())
	}
	logger.Debug("Done initializing store")

	logger.Debug("Registering backend...")
	backendID := etcd.NewBackendIDGetter(b.ctx, client)
	logger.Debug("Done registering backend.")

	// Initialize an etcd getter
	queueGetter := queue.EtcdGetter{Client: client, BackendIDGetter: backendID}

	// Initialize the bus
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{
		RingGetter: ring.EtcdGetter{Client: client, BackendID: fmt.Sprintf("%x", backendID.GetBackendID())},
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
		Host:        config.APIHost,
		Port:        config.APIPort,
		Bus:         bus,
		Store:       store,
		QueueGetter: queueGetter,
		TLS:         config.TLS,
		Cluster:     clientv3.NewCluster(client),
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

	// Add store to our backend, since it's needed across the methods
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

	if b.Etcd != nil {
		// Add etcd to our errGroup, since it's not included in the daemon list
		eg.errors = append(eg.errors, b.Etcd)
	}
	eg.Go()

	select {
	case err := <-eg.Err():
		logger.Error(err.Error())
	case <-b.ctx.Done():
		logger.Info("backend shutting down")
	}

	var derr error

	if err := sg.Stop(); err != nil {
		if derr == nil {
			derr = err
		}
	}

	if b.Etcd != nil {
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

// Stop the Backend cleanly.
func (b *Backend) Stop() {
	b.cancel()
	<-b.done
}
