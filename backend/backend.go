package backend

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sensu/sensu-go/backend/apid"
	"github.com/sensu/sensu-go/backend/daemon"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/pipelined"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
)

var (
	// upgrader is safe for concurrent use, and we don't need any particularly
	// specialized configurations for different uses.
	upgrader = &websocket.Upgrader{}
)

// Config specifies a Backend configuration.
type Config struct {
	APIPort             int
	AgentPort           int
	StateDir            string
	EtcdClientListenURL string
	EtcdPeerListenURL   string
	EtcdInitialCluster  string
}

// A Backend is a Sensu Backend server responsible for handling incoming
// HTTP requests and upgrading them
type Backend struct {
	Config *Config

	shutdownChan   chan struct{}
	done           chan struct{}
	messageBus     messaging.MessageBus
	apid           daemon.Daemon
	agentServer    *http.Server
	checkScheduler *Checker
	etcd           *etcd.Etcd

	pipelined daemon.Daemon
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

	asrv, err := agentServer(b)
	if err != nil {
		return nil, err
	}

	b.agentServer = asrv

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

	b.checkScheduler = &Checker{
		MessageBus: b.messageBus,
		Client:     cli,
		Store:      st,
	}
	err = b.checkScheduler.Start()
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

	b.apid = &apid.APId{
		Store:         st,
		Port:          b.Config.APIPort,
		BackendStatus: b.Status,
	}
	if err := b.apid.Start(); err != nil {
		return err
	}

	// there are two channels in play here: inErrChan is used by the various
	// services the Backend manages as a destination for terminal errors.
	// we then monitor that channel and the first error returned on that
	// channel causes the Backend to shutdown. Then, we pass that error on
	// to the Err() method via the b.errChan channel.
	inErrChan := make(chan error)

	go func() {
		inErrChan <- <-b.apid.Err()
	}()

	go func() {
		log.Println("starting agent server on address: ", b.agentServer.Addr)
		inErrChan <- b.agentServer.ListenAndServe()
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

	select {
	case err := <-inErrChan:
		log.Println(err.Error())
	case <-b.shutdownChan:
		log.Println("backend shutting down")
	}

	log.Printf("shutting down etcd")
	if err := b.etcd.Shutdown(); err != nil {
		log.Printf("error shutting down etcd: %s", err.Error())
	}
	log.Printf("shutting down apid")
	b.apid.Stop()
	log.Printf("shutting down message bus")
	b.messageBus.Stop()
	log.Printf("shutting down pipelined")
	b.pipelined.Stop()

	// we allow inErrChan to leak to avoid panics from other
	// goroutines writing errors to either after shutdown has been initiated.
	close(b.done)

	return nil
}

// Status returns a map of component name to boolean healthy indicator.
func (b *Backend) Status() types.StatusMap {
	sm := map[string]bool{
		"store":       b.etcd.Healthy(),
		"message_bus": true,
		"pipelined":   true,
		"apid":        true,
	}

	if b.messageBus.Status() != nil {
		sm["message_bus"] = false
	}

	if b.pipelined.Status() != nil {
		sm["pipelined"] = false
	}

	if b.apid.Status() != nil {
		sm["apid"] = false
	}

	return sm
}

// Stop the Backend cleanly.
func (b *Backend) Stop() {
	close(b.shutdownChan)
	<-b.done
}

func agentServer(b *Backend) (*http.Server, error) {
	store, err := b.etcd.NewStore()
	if err != nil {
		return nil, err
	}

	r := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("transport error on websocket upgrade: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session, err := NewSession(transport.NewTransport(conn), b.messageBus, store)
		if err != nil {
			log.Println("failed to start session: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = session.Start()
		if err != nil {
			log.Println("failed to start session: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", b.Config.AgentPort),
		Handler:      r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}, nil
}
