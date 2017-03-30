package backend

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sensu/sensu-go/backend/messaging"
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
	messageBus   messaging.MessageBus
	httpServer   *http.Server
	agentServer  *http.Server
	etcd         *etcd.Etcd
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

	httpsrv, err := httpServer(b)
	if err != nil {
		return nil, err
	}
	b.httpServer = httpsrv

	asrv, err := agentServer(b)
	if err != nil {
		return nil, err
	}

	b.agentServer = asrv

	b.messageBus = &messaging.MemoryBus{}

	return b, nil
}

// Run starts all of the Backend server's event loops and sets up the HTTP
// server.
func (b *Backend) Run() error {
	if err := b.messageBus.Start(); err != nil {
		return err
	}

	// there are two channels in play here: inErrChan is used by the various
	// services the Backend manages as a destination for terminal errors.
	// we then monitor that channel and the first error returned on that
	// channel causes the Backend to shutdown. Then, we pass that error on
	// to the Err() method via the b.errChan channel.
	inErrChan := make(chan error)

	go func() {
		inErrChan <- b.httpServer.ListenAndServe()
	}()

	go func() {
		inErrChan <- b.agentServer.ListenAndServe()
	}()

	go func() {
		inErrChan <- <-b.etcd.Err()
	}()

	go func() {
		inErrChan <- <-b.messageBus.Err()
	}()

	go func() {
		var inErr error
		select {
		case inErr = <-inErrChan:
			log.Fatal("http server error: ", inErr.Error())
		case <-b.shutdownChan:
			log.Println("backend shutting down")
		}

		log.Printf("shutting down etcd")
		if err := b.etcd.Shutdown(); err != nil {
			log.Printf("error shutting down etcd: %s", err.Error())
		}
		log.Printf("shutting down http server")
		if err := b.httpServer.Shutdown(context.TODO()); err != nil {
			log.Printf("error shutting down http listener: %s", err.Error())
		}
		log.Printf("shutting down message bus")
		b.messageBus.Stop()

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
		"store":       b.etcd.Healthy(),
		"message_bus": true,
	}

	if b.messageBus.Status() != nil {
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

func agentServer(b *Backend) (*http.Server, error) {
	store, err := b.etcd.NewStore()
	if err != nil {
		return nil, err
	}

	r := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("transport error on websocket upgrade: ", err.Error())
			return
		}

		session := NewSession(transport.NewTransport(conn), b.messageBus, store)
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
	}, nil
}
