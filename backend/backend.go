package backend

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"

	"github.com/nsqio/nsq/nsqd"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/transport"
)

// Config specifies a Backend configuration.
type Config struct {
	Port                int
	StateDir            string
	EtcdClientListenURL string
	EtcdPeerListenURL   string
	EtcdInitialCluster  string
}

// A Backend is a Sensu Backend server responsible for handling incoming
// HTTP requests and upgrading them
type Backend struct {
	Config *Config

	errChan         chan error
	shutdownChan    chan struct{}
	done            chan struct{}
	messageBus      *nsqd.NSQD
	httpServer      *http.Server
	transportServer *transport.Server
}

// NewBackend will, given a Config, create an initialized Backend and return a
// pointer to it.
func NewBackend(config *Config) (*Backend, error) {
	b := &Backend{
		Config: config,

		done:         make(chan struct{}),
		errChan:      make(chan error, 1),
		shutdownChan: make(chan struct{}),
	}
	b.httpServer = b.newHTTPServer()
	b.transportServer = transport.NewServer()
	nsqConfig := messaging.NewConfig()
	nsqConfig.StatePath = filepath.Join(config.StateDir, "nsqd")
	bus, err := messaging.NewNSQD(nsqConfig)
	if err != nil {
		return nil, err
	}
	b.messageBus = bus
	return b, nil
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
	cfg := etcd.NewConfig()
	cfg.StateDir = b.Config.StateDir
	cfg.ClientListenURL = b.Config.EtcdClientListenURL
	cfg.PeerListenURL = b.Config.EtcdPeerListenURL
	cfg.InitialCluster = b.Config.EtcdInitialCluster
	e, err := etcd.NewEtcd(cfg)
	if err != nil {
		return fmt.Errorf("error starting etcd: %s", err.Error())
	}

	go func() {
		errChan <- b.httpServer.ListenAndServe()
	}()
	go func() {
		errChan <- <-e.Err()
	}()
	go b.messageBus.Main()

	go func() {
		var inErr error
		select {
		case inErr = <-errChan:
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
		close(b.shutdownChan)
		close(b.errChan)
		close(b.done)
		cmd := exec.Command("nslookup", "-l", "-n", "-t")
		cmd.Run()
		rdr, _ := cmd.StdoutPipe()
		out, _ := ioutil.ReadAll(io.Reader(rdr))
		fmt.Print(out)
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
	<-b.done
}
