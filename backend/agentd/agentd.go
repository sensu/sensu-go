package agentd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/transport"
)

var (
	// upgrader is safe for concurrent use, and we don't need any particularly
	// specialized configurations for different uses.
	upgrader = &websocket.Upgrader{}
)

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
	store      store.Store
	bus        messaging.MessageBus
	tls        *corev2.TLSOptions
	ringPool   *ringv2.Pool
}

// Config configures an Agentd.
type Config struct {
	Host     string
	Port     int
	Bus      messaging.MessageBus
	Store    store.Store
	TLS      *corev2.TLSOptions
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

	handler := middlewares.BasicAuthentication(middlewares.BasicAuthorization(http.HandlerFunc(a.webSocketHandler), a.store), a.store)
	a.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.Host, a.Port),
		Handler:      handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		TLSConfig:    tlsServerConfig,
		// Capture the log entries from agentd's HTTP server
		ErrorLog: log.New(&logrusIOWriter{entry: logger}, "", 0),
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
	var marshal MarshalFunc
	var unmarshal UnmarshalFunc
	var contentType string
	responseHeader := make(http.Header)
	responseHeader.Add("Accept", ProtobufSerializationHeader)
	logger.WithField("header", fmt.Sprintf("Accept: %s", ProtobufSerializationHeader)).Debug("setting header")
	responseHeader.Add("Accept", JSONSerializationHeader)
	logger.WithField("header", fmt.Sprintf("Accept: %s", JSONSerializationHeader)).Debug("setting header")
	if r.Header.Get("Accept") == ProtobufSerializationHeader {
		marshal = proto.Marshal
		unmarshal = proto.Unmarshal
		contentType = ProtobufSerializationHeader
		logger.WithField("format", "protobuf").Debug("setting serialization/deserialization")
	} else {
		marshal = MarshalJSON
		unmarshal = UnmarshalJSON
		contentType = JSONSerializationHeader
		logger.WithField("format", "JSON").Debug("setting serialization/deserialization")
	}
	responseHeader.Set("Content-Type", contentType)
	logger.WithField("header", fmt.Sprintf("Content-Type: %s", contentType)).Debug("setting header")

	conn, err := upgrader.Upgrade(w, r, responseHeader)
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
		ContentType:   contentType,
	}

	cfg.Subscriptions = addEntitySubscription(cfg.AgentName, cfg.Subscriptions)

	session, err := NewSession(cfg, transport.NewTransport(conn), a.bus, a.store, unmarshal, marshal)
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
