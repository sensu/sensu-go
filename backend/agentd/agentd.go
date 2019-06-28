package agentd

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
)

var (
	// upgrader is safe for concurrent use, and we don't need any particularly
	// specialized configurations for different uses.
	upgrader = &websocket.Upgrader{}
)

// Store specifies storage requirements for Agentd.
type Store interface {
	middlewares.AuthStore
	SessionStore
}

// Agentd is the backend HTTP API.
type Agentd struct {
	// Host is the hostname Agentd is running on.
	Host string

	// Port is the port Agentd is running on.
	Port int

	stopping   chan struct{}
	running    *atomic.Value
	wg         *sync.WaitGroup
	errChan    chan error
	httpServer *http.Server
	store      Store
	bus        messaging.MessageBus
	tls        *types.TLSOptions
	ringPool   *ringv2.Pool
}

// Config configures an Agentd.
type Config struct {
	Host     string
	Port     int
	Bus      messaging.MessageBus
	Store    store.Store
	TLS      *types.TLSOptions
	RingPool *ringv2.Pool
}

// Option is a functional option.
type Option func(*Agentd) error

// New creates a new Agentd.
func New(c Config, opts ...Option) (*Agentd, error) {
	a := &Agentd{
		Host:     c.Host,
		Port:     c.Port,
		bus:      c.Bus,
		store:    c.Store,
		tls:      c.TLS,
		stopping: make(chan struct{}, 1),
		running:  &atomic.Value{},
		wg:       &sync.WaitGroup{},
		errChan:  make(chan error, 1),
		ringPool: c.RingPool,
	}

	// prepare server TLS config
	tlsServerConfig, err := c.TLS.ToServerTLSConfig()
	if err != nil {
		return nil, err
	}

	handler := middlewares.BasicAuthentication(http.HandlerFunc(a.webSocketHandler), a.store)
	a.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.Host, a.Port),
		Handler:      handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		TLSConfig:    tlsServerConfig,
	}
	for _, o := range opts {
		if err := o(a); err != nil {
			return nil, err
		}
	}
	return a, nil
}

// Start Agentd.
func (a *Agentd) Start() error {
	logger.Info("starting agentd on address: ", a.httpServer.Addr)
	a.wg.Add(1)

	go func() {
		defer a.wg.Done()
		var err error
		if a.tls != nil {
			// TLS configuration comes from ToServerTLSConfig
			err = a.httpServer.ListenAndServeTLS("", "")
		} else {
			err = a.httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Error("failed to start http/https server")
		}
	}()

	_ = prometheus.Register(sessionCounter)

	return nil
}

// Stop Agentd.
func (a *Agentd) Stop() error {
	if err := a.httpServer.Shutdown(context.TODO()); err != nil {
		// failure/timeout shutting down the server gracefully
		logger.Error("failed to shutdown http server gracefully - forcing shutdown")
		if closeErr := a.httpServer.Close(); closeErr != nil {
			logger.Error("failed to shutdown http server forcefully")
		}
	}
	a.running.Store(false)
	close(a.stopping)
	a.wg.Wait()
	close(a.errChan)

	return nil
}

// Err returns a channel to listen for terminal errors on.
func (a *Agentd) Err() <-chan error {
	return a.errChan
}

// Name returns the daemon name
func (a *Agentd) Name() string {
	return "agentd"
}

func (a *Agentd) webSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.WithField("addr", r.RemoteAddr).WithError(err).Error("transport error on websocket upgrade")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cfg := SessionConfig{
		AgentAddr:     r.RemoteAddr,
		AgentName:     r.Header.Get(transport.HeaderKeyAgentName),
		Namespace:     r.Header.Get(transport.HeaderKeyNamespace),
		User:          r.Header.Get(transport.HeaderKeyUser),
		Subscriptions: strings.Split(r.Header.Get(transport.HeaderKeySubscriptions), ","),
		RingPool:      a.ringPool,
	}

	cfg.Subscriptions = addEntitySubscription(cfg.AgentName, cfg.Subscriptions)

	session, err := NewSession(cfg, transport.NewTransport(conn), a.bus, a.store)
	if err != nil {
		logger.WithError(err).Error("failed to create session")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = session.Start()
	if err != nil {
		logger.WithError(err).Error("failed to start session")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
