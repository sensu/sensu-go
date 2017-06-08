package backend

import (
	"fmt"

	"github.com/gorilla/websocket"

	"github.com/sensu/sensu-go/backend/agentd"
	"github.com/sensu/sensu-go/backend/apid"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authentication/providers/basic"
	"github.com/sensu/sensu-go/backend/daemon"
	"github.com/sensu/sensu-go/backend/dashboardd"
	"github.com/sensu/sensu-go/backend/eventd"
	"github.com/sensu/sensu-go/backend/keepalived"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/pipelined"
	"github.com/sensu/sensu-go/backend/schedulerd"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/types"
)

var (
	// upgrader is safe for concurrent use, and we don't need any particularly
	// specialized configurations for different uses.
	upgrader = &websocket.Upgrader{}
)

// Config specifies a Backend configuration.
type Config struct {
	AgentHost             string
	AgentPort             int
	APIAuthentication     bool
	APIHost               string
	APIPort               int
	DashboardDir          string
	DashboardHost         string
	DashboardPort         int
	DeregistrationHandler string
	StateDir              string
	EtcdClientListenURL   string
	EtcdPeerListenURL     string
	EtcdInitialCluster    string
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
	if config.EtcdClientListenURL == "" {
		config.EtcdClientListenURL = "http://127.0.0.1:2379"
	}

	if config.EtcdPeerListenURL == "" {
		config.EtcdPeerListenURL = "http://127.0.0.1:2380"
	}

	if config.EtcdInitialCluster == "" {
		config.EtcdInitialCluster = "default=http://127.0.0.1:2380"
	}

	if config.APIPort == 0 {
		config.APIPort = 8080
	}

	if config.AgentPort == 0 {
		config.AgentPort = 8081
	}

	b := &Backend{
		Config: config,

		done:         make(chan struct{}),
		shutdownChan: make(chan struct{}),
	}

	// we go ahead and setup and start etcd here, because we'll have to pass
	// a store along to the API.
	cfg := etcd.NewConfig()
	cfg.StateDir = b.Config.StateDir
	cfg.ClientListenURL = b.Config.EtcdClientListenURL
	cfg.PeerListenURL = b.Config.EtcdPeerListenURL
	cfg.InitialCluster = b.Config.EtcdInitialCluster
	e, err := etcd.NewEtcd(cfg)
	if err != nil {
		return nil, fmt.Errorf("error starting etcd: %s", err.Error())
	}
	b.etcd = e

	b.messageBus = &messaging.WizardBus{}

	return b, nil
}

// Run starts all of the Backend server's event loops and sets up the HTTP
// server.
func (b *Backend) Run() error {
	if err := b.messageBus.Start(); err != nil {
		return err
	}

	// Right now, instantiating a new Etcd will start etcd. If we change that
	// s.t. Etcd has its own Start() method, conforming to Daemon, then we will
	// want to make sure that we aren't calling NewClient before starting it,
	// I think. That might return a connection error.
	st, err := b.etcd.NewStore()
	if err != nil {
		return err
	}

	cli, err := b.etcd.NewClient()
	if err != nil {
		return err
	}

	b.schedulerd = &schedulerd.Schedulerd{
		MessageBus: b.messageBus,
		Store:      st,
	}
	err = b.schedulerd.Start()
	if err != nil {
		return err
	}

	b.pipelined = &pipelined.Pipelined{
		Store:      st,
		MessageBus: b.messageBus,
	}
	if err := b.pipelined.Start(); err != nil {
		return err
	}

	// Initializes the JWT secret
	jwt.InitSecret(st)

	// TODO(Simon): We need to determine the authentication driver from the config
	auth := &basic.Basic{
		Enabled: b.Config.APIAuthentication,
		Store:   st,
	}

	b.apid = &apid.APId{
		Authentication: auth,
		Store:          st,
		Host:           b.Config.APIHost,
		Port:           b.Config.APIPort,
		BackendStatus:  b.Status,
	}
	if err := b.apid.Start(); err != nil {
		return err
	}

	b.agentd = &agentd.Agentd{
		Store:      st,
		Host:       b.Config.AgentHost,
		Port:       b.Config.AgentPort,
		MessageBus: b.messageBus,
	}
	if err := b.agentd.Start(); err != nil {
		return err
	}

	b.dashboardd = &dashboardd.Dashboardd{
		BackendStatus: b.Status,
		Config: dashboardd.Config{
			Dir:  b.Config.DashboardDir,
			Host: b.Config.DashboardHost,
			Port: b.Config.DashboardPort,
		},
	}
	if err := b.dashboardd.Start(); err != nil {
		return err
	}

	b.eventd = &eventd.Eventd{
		Store:      st,
		MessageBus: b.messageBus,
	}
	if err := b.eventd.Start(); err != nil {
		return err
	}

	b.keepalived = &keepalived.Keepalived{
		Store:                 st,
		MessageBus:            b.messageBus,
		DeregistrationHandler: b.Config.DeregistrationHandler,
	}
	if err := b.keepalived.Start(); err != nil {
		return err
	}

	inErrChan := make(chan error)

	go func() {
		inErrChan <- <-b.apid.Err()
	}()

	go func() {
		inErrChan <- <-b.agentd.Err()
	}()

	go func() {
		inErrChan <- <-b.schedulerd.Err()
	}()

	go func() {
		inErrChan <- <-b.etcd.Err()
	}()

	go func() {
		inErrChan <- <-b.messageBus.Err()
	}()

	go func() {
		inErrChan <- <-b.pipelined.Err()
	}()

	go func() {
		inErrChan <- <-b.dashboardd.Err()
	}()

	go func() {
		inErrChan <- <-b.eventd.Err()
	}()

	go func() {
		inErrChan <- <-b.keepalived.Err()
	}()

	select {
	case err := <-inErrChan:
		logger.Error(err.Error())
	case <-b.shutdownChan:
		logger.Info("backend shutting down")
	}

	logger.Info("shutting down etcd")
	if err := b.etcd.Shutdown(); err != nil {
		logger.Errorf("error shutting down etcd: %s", err.Error())
	}

	// stop allowing API connections
	logger.Info("shutting down apid")
	b.apid.Stop()

	// stop allowing dashboard connections
	logger.Info("shutting down dashboardd")
	b.dashboardd.Stop()

	// disconnect all agents and don't allow any more to connect.
	logger.Info("shutting down agentd")
	b.agentd.Stop()

	// stop scheduling checks.
	logger.Info("shutting down schedulerd")
	b.schedulerd.Stop()

	// Shutting down eventd will cause it to drain events to the bus
	logger.Info("shutting down eventd")
	b.eventd.Stop()

	// Once events have been drained from eventd, pipelined can finish
	// processing events.
	logger.Info("shutting down pipelined")
	b.pipelined.Stop()

	// finally shutdown the message bus once all other components have stopped
	// using it.
	logger.Info("shutting down message bus")
	b.messageBus.Stop()

	// we allow inErrChan to leak to avoid panics from other
	// goroutines writing errors to either after shutdown has been initiated.
	close(b.done)

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
