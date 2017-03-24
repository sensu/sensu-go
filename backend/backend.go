package backend

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nsqio/nsq/nsqd"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/transport"
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

	errChan      chan error
	shutdownChan chan struct{}
	done         chan struct{}
	messageBus   *nsqd.NSQD
	httpServer   *http.Server
	agentServer  *http.Server
	store        store.Store
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
		errChan:      make(chan error, 1),
		shutdownChan: make(chan struct{}),
	}

	b.httpServer = httpServer(b)
	b.agentServer = agentServer(b)
	nsqConfig := messaging.NewConfig()
	nsqConfig.StatePath = filepath.Join(config.StateDir, "nsqd")
	bus, err := messaging.NewNSQD(nsqConfig)
	if err != nil {
		return nil, err
	}
	b.messageBus = bus

	return b, nil
}

// Run starts all of the Backend server's event loops and sets up the HTTP
// server.
func (b *Backend) Run() error {
	// there are two channels in play here: inErrChan is used by the various
	// services the Backend manages as a destination for terminal errors.
	// we then monitor that channel and the first error returned on that
	// channel causes the Backend to shutdown. Then, we pass that error on
	// to the Err() method via the b.errChan channel.
	inErrChan := make(chan error)
	cfg := etcd.NewConfig()
	cfg.StateDir = b.Config.StateDir
	cfg.ClientListenURL = b.Config.EtcdClientListenURL
	cfg.PeerListenURL = b.Config.EtcdPeerListenURL
	cfg.InitialCluster = b.Config.EtcdInitialCluster
	e, err := etcd.NewEtcd(cfg)
	if err != nil {
		return fmt.Errorf("error starting etcd: %s", err.Error())
	}

	store, err := e.NewStore()
	if err != nil {
		return err
	}
	b.store = store

	go func() {
		inErrChan <- b.httpServer.ListenAndServe()
	}()

	go func() {
		inErrChan <- b.agentServer.ListenAndServe()
	}()

	go func() {
		inErrChan <- <-e.Err()
	}()

	go b.messageBus.Main()

	go func() {
		var inErr error
		select {
		case inErr = <-inErrChan:
			log.Fatal("http server error: ", inErr.Error())
		case <-b.shutdownChan:
			log.Println("backend shutting down")
		}

		log.Printf("shutting down etcd")
		if err := e.Shutdown(); err != nil {
			log.Printf("error shutting down etcd: %s", err.Error())
		}
		log.Printf("shutting down http server")
		if err := b.httpServer.Shutdown(context.TODO()); err != nil {
			log.Printf("error shutting down http listener: %s", err.Error())
		}
		log.Printf("shutting down message bus")
		b.messageBus.Exit()

		// if an error caused the shutdown
		if inErr != nil {
			b.errChan <- inErr
		}
		// we allow b.errChan and inErrChan to leak to avoid panics from other
		// goroutines writing errors to either after shutdown has been initiated.
		close(b.done)
	}()

	return nil
}

// Status returns a map of component name to boolean healthy indicator.
func (b *Backend) Status() StatusMap {
	sm := map[string]bool{
		"store":       b.store.Healthy(),
		"message_bus": true,
	}

	busHealth := b.messageBus.GetHealth()
	// ugh.
	if busHealth != "OK" {
		sm["message_bus"] = false
	}

	return sm
}

// Err blocks and returns the first terminal error encountered by this Backend.
func (b *Backend) Err() error {
	return <-b.errChan
}

// Stop the Backend cleanly.
func (b *Backend) Stop() {
	close(b.shutdownChan)
	<-b.done
}

func agentServer(b *Backend) *http.Server {
	r := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("transport error on websocket upgrade: ", err.Error())
			return
		}

		session := NewSession(transport.NewTransport(conn), b.store)
		err = session.Start()
		if err != nil {
			log.Println("failed to start session: ", err.Error())
		}
	})

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", b.Config.AgentPort),
		Handler:      r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}
