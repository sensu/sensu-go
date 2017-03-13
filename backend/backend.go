package backend

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/nsqio/nsq/nsqd"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/transport"
)

// Config specifies a Backend configuration.
type Config struct {
	Port int
}

// A Backend is a Sensu Backend server responsible for handling incoming
// HTTP requests and upgrading them
type Backend struct {
	Config *Config

	errChan         chan error
	shutdownChan    chan struct{}
	messageBus      *nsqd.NSQD
	httpServer      *http.Server
	transportServer *transport.Server
}

// NewBackend will, given a Config, create an initialized Backend and return a
// pointer to it.
func NewBackend(config *Config) *Backend {
	b := &Backend{
		Config: config,

		errChan:      make(chan error, 1),
		shutdownChan: make(chan struct{}),
	}
	b.httpServer = b.newHTTPServer()
	b.transportServer = transport.NewServer()
	b.messageBus = messaging.NewNSQD()
	return b
}

func (b *Backend) newHTTPServer() *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", b.Config.Port),
		Handler: b.newHTTPHandler(),
	}
}

func (b *Backend) newHTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := b.transportServer.Serve(w, r)
		if err != nil {
			log.Println("transport error on websocket upgrade: ", err.Error())
			return
		}

		session := NewSession(conn)
		// blocks until session end is detected
		err = session.Start()
		if err != nil {
			log.Println("failed to start session: ", err.Error())
		}
	})
}

// Run starts all of the Backend server's event loops and sets up the HTTP
// server.
func (b *Backend) Run() error {
	errChan := make(chan error)
	if err := etcd.NewEtcd(etcd.NewConfig()); err != nil {
		return fmt.Errorf("error starting etcd: %s", err.Error())
	}

	go func() {
		errChan <- b.httpServer.ListenAndServe()
	}()
	go func() {
		errChan <- <-etcd.Err()
	}()

	go func() {
		var inErr error
		select {
		case inErr = <-errChan:
			log.Fatal("http server error: ", inErr.Error())
		case <-b.shutdownChan:
			log.Println("backend shutting down")
		}

		if err := etcd.Shutdown(); err != nil {
			log.Printf("error shutting down etcd: %s", err.Error())
		}
		if err := b.httpServer.Shutdown(context.TODO()); err != nil {
			log.Printf("error shutting down http listener: %s", err.Error())
		}

		// if an error caused the shutdown
		if inErr != nil {
			b.errChan <- inErr
		}
		close(b.errChan)
	}()

	return nil
}

// Err returns a channel yielding terminal errors for the backend. The channel
// will be closed on clean shutdown.
func (b *Backend) Err() <-chan error {
	return b.errChan
}

// Stop the Backend cleanly.
func (b *Backend) Stop() {
	// TODO(greg): shutdown all active client connections
	b.shutdownChan <- struct{}{}
}
