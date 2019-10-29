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
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
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
	ctx        context.Context
	cancel     context.CancelFunc
}

// Config configures an Agentd.
type Config struct {
	Host         string
	Port         int
	Bus          messaging.MessageBus
	Store        store.Store
	TLS          *corev2.TLSOptions
	RingPool     *ringv2.Pool
	WriteTimeout int
}

// Option is a functional option.
type Option func(*Agentd) error

// New creates a new Agentd.
func New(c Config, opts ...Option) (*Agentd, error) {
	ctx, cancel := context.WithCancel(context.Background())
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
		ctx:      ctx,
		cancel:   cancel,
	}

	// prepare server TLS config
	tlsServerConfig, err := c.TLS.ToServerTLSConfig()
	if err != nil {
		return nil, err
	}

	// Configure the middlewares used by agentd's HTTP server by assigning them to
	// public variables so they can be overriden from the enterprise codebase
	AuthenticationMiddleware = a.AuthenticationMiddleware
	AuthorizationMiddleware = a.AuthorizationMiddleware
	AgentLimiterMiddleware = a.AgentLimiterMiddleware
	EntityLimiterMiddleware = a.EntityLimiterMiddleware

	// Initialize a mux router that indirectly uses our middlewares defined above.
	// We can't directly use them because mux will keep a copy of the middleware
	// functions, which prevent us from modifying the actual middleware logic at
	// runtime, so we need this workaround
	router := mux.NewRouter()
	router.HandleFunc("/", a.webSocketHandler)
	router.Use(authenticate, authorize, entityLimit, agentLimit)

	a.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.Host, a.Port),
		Handler:      router,
		WriteTimeout: time.Duration(c.WriteTimeout) * time.Second,
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
	a.cancel()
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

	session, err := NewSession(a.ctx, cfg, transport.NewTransport(conn), a.bus, a.store, unmarshal, marshal)
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

// AuthenticationMiddleware represents the core authentication middleware for
// agentd, which consists of basic authentication.
func (a *Agentd) AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "missing credentials", http.StatusUnauthorized)
			return
		}

		// Authenticate against the provider
		user, err := a.store.AuthenticateUser(r.Context(), username, password)
		if err != nil {
			logger.
				WithField("user", username).
				WithError(err).
				Error("invalid username and/or password")
			http.Error(w, "bad credentials", http.StatusUnauthorized)
			return
		}

		// The user was authenticated against the local store, therefore add the
		// system:user group so it can view itself and change its password
		user.Groups = append(user.Groups, "system:user")

		// TODO: eventually break out authorization details in context from jwt
		// claims; in this method they are too tightly bound
		claims, _ := jwt.NewClaims(user)
		ctx := jwt.SetClaimsIntoContext(r, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthorizationMiddleware represents the core authorization middleware for
// agentd, which consists of making sure the agent's entity is authorized to
// create events in the given namespace
func (a *Agentd) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		namespace := r.Header.Get(transport.HeaderKeyNamespace)
		ctx := r.Context()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, namespace)

		attrs := &authorization.Attributes{
			APIGroup:     "core",
			APIVersion:   "v2",
			Namespace:    namespace,
			Resource:     "events",
			ResourceName: r.Header.Get(transport.HeaderKeyAgentName),
			Verb:         "create",
		}

		err := middlewares.GetUser(ctx, attrs)
		if err != nil {
			logger.WithError(err).Info("could not get user from request context")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		auth := &rbac.Authorizer{
			Store: a.store,
		}
		authorized, err := auth.Authorize(ctx, attrs)
		if err != nil {
			logger.WithError(err).Error("unexpected error while authorization the session")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		if !authorized {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AgentLimiterMiddleware is a placeholder middleware for the agent
// HTTP session. It will be overwritten by an enterprise middleware.
func (a *Agentd) AgentLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// EntityLimiterMiddleware is a placeholder middleware for the agent
// HTTP session. It will be overwritten by an enterprise middleware.
func (a *Agentd) EntityLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
