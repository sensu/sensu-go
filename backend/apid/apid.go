package apid

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
	"github.com/sensu/sensu-go/backend/apid/routers"
	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// APId is the backend HTTP API.
type APId struct {
	Authenticator    *authentication.Authenticator
	HTTPServer       *http.Server
	CoreSubrouter    *mux.Router
	GraphQLSubrouter *mux.Router

	stopping            chan struct{}
	running             *atomic.Value
	wg                  *sync.WaitGroup
	errChan             chan error
	bus                 messaging.MessageBus
	store               store.Store
	eventStore          store.EventStore
	queueGetter         types.QueueGetter
	tls                 *types.TLSOptions
	cluster             clientv3.Cluster
	etcdClientTLSConfig *tls.Config
	clusterVersion      string
}

// Option is a functional option.
type Option func(*APId) error

// Config configures APId.
type Config struct {
	ListenAddress       string
	URL                 string
	Bus                 messaging.MessageBus
	Store               store.Store
	EventStore          store.EventStore
	QueueGetter         types.QueueGetter
	TLS                 *types.TLSOptions
	Cluster             clientv3.Cluster
	EtcdClientTLSConfig *tls.Config
	Authenticator       *authentication.Authenticator
	ClusterVersion      string
}

// New creates a new APId.
func New(c Config, opts ...Option) (*APId, error) {
	a := &APId{
		store:               c.Store,
		eventStore:          c.EventStore,
		queueGetter:         c.QueueGetter,
		tls:                 c.TLS,
		bus:                 c.Bus,
		stopping:            make(chan struct{}, 1),
		running:             &atomic.Value{},
		wg:                  &sync.WaitGroup{},
		errChan:             make(chan error, 1),
		cluster:             c.Cluster,
		etcdClientTLSConfig: c.EtcdClientTLSConfig,
		Authenticator:       c.Authenticator,
		clusterVersion:      c.ClusterVersion,
	}

	// prepare TLS config
	var tlsServerConfig *tls.Config
	var err error
	if c.TLS != nil {
		tlsServerConfig, err = c.TLS.ToServerTLSConfig()
		if err != nil {
			return nil, err
		}
	}

	router, coreSubrouter, graphQLSubrouter, err := NewHandler(c)
	if err != nil {
		return nil, err
	}

	a.CoreSubrouter = coreSubrouter
	a.GraphQLSubrouter = graphQLSubrouter

	a.HTTPServer = &http.Server{
		Addr:         c.ListenAddress,
		Handler:      router,
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

func NewHandler(c Config) (root http.Handler, core *mux.Router, graphql *mux.Router, err error) {
	router := mux.NewRouter().UseEncodedPath()
	router.NotFoundHandler = middlewares.SimpleLogger{}.Then(http.HandlerFunc(notFoundHandler))
	router.Handle("/metrics", promhttp.Handler())
	registerUnauthenticatedResources(router, c.Store, c.Cluster, c.EtcdClientTLSConfig, c.ClusterVersion, c.Bus)
	graphQLSubrouter := NewSubrouter(
		router.NewRoute(),
		middlewares.SimpleLogger{},
		middlewares.LimitRequest{},
		// TODO: Currently the web app relies on receiving a 401 to determine if
		//       a user is not authenticated. However, in the future we should
		//       allow requests without an access token to continue so that
		//       unauthenticated clients can still fetch the schema. Useful for
		//       implementing tools like GraphiQL.
		//
		//       https://github.com/graphql/graphiql
		//       https://graphql.org/learn/introspection/
		middlewares.Authentication{IgnoreUnauthorized: false},
		middlewares.AllowList{Store: c.Store, IgnoreMissingClaims: true},
	)
	mountRouters(
		graphQLSubrouter,
		routers.NewGraphQLRouter(c.Store, c.EventStore, &rbac.Authorizer{Store: c.Store}, c.QueueGetter, c.Bus),
	)
	registerAuthenticationResources(router, c.Store, c.Authenticator)
	coreSubrouter := NewSubrouter(
		router.NewRoute().
			PathPrefix("/api/{group:core}/{version:v2}/"),
		middlewares.SimpleLogger{},
		middlewares.Namespace{},
		middlewares.Authentication{},
		middlewares.AllowList{Store: c.Store},
		middlewares.AuthorizationAttributes{},
		middlewares.Authorization{Authorizer: &rbac.Authorizer{Store: c.Store}},
		middlewares.LimitRequest{},
		middlewares.Pagination{},
	)
	mountRouters(
		coreSubrouter,
		routers.NewAssetRouter(c.Store),
		routers.NewChecksRouter(c.Store, c.QueueGetter),
		routers.NewClusterRolesRouter(c.Store),
		routers.NewClusterRoleBindingsRouter(c.Store),
		routers.NewClusterRouter(actions.NewClusterController(c.Cluster, c.Store)),
		routers.NewEntitiesRouter(c.Store, c.EventStore),
		routers.NewEventFiltersRouter(c.Store),
		routers.NewEventsRouter(c.EventStore, c.Bus),
		routers.NewExtensionsRouter(c.Store),
		routers.NewHandlersRouter(c.Store),
		routers.NewHooksRouter(c.Store),
		routers.NewMutatorsRouter(c.Store),
		routers.NewNamespacesRouter(c.Store),
		routers.NewRolesRouter(c.Store),
		routers.NewRoleBindingsRouter(c.Store),
		routers.NewSilencedRouter(c.Store),
		routers.NewTessenRouter(actions.NewTessenController(c.Store, c.Bus)),
		routers.NewUsersRouter(c.Store),
	)

	return router, coreSubrouter, graphQLSubrouter, nil
}

func notFoundHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	resp := map[string]interface{}{
		"message": "not found", "code": actions.NotFound,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// Start APId.
func (a *APId) Start() error {
	logger.Info("starting apid on address: ", a.HTTPServer.Addr)
	a.wg.Add(1)

	go func() {
		defer a.wg.Done()
		var err error
		if a.tls != nil {
			// TLS configuration comes from ToServerTLSConfig
			err = a.HTTPServer.ListenAndServeTLS("", "")
		} else {
			err = a.HTTPServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			a.errChan <- fmt.Errorf("failed to start http/https server %s", err)
		}
	}()

	return nil
}

// Stop httpApi.
func (a *APId) Stop() error {
	if err := a.HTTPServer.Shutdown(context.TODO()); err != nil {
		// failure/timeout shutting down the server gracefully
		logger.Error("failed to shutdown http server gracefully - forcing shutdown")
		if closeErr := a.HTTPServer.Close(); closeErr != nil {
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
func (a *APId) Err() <-chan error {
	return a.errChan
}

// Name returns the daemon name
func (a *APId) Name() string {
	return "apid"
}

func registerUnauthenticatedResources(
	router *mux.Router,
	store store.Store,
	cluster clientv3.Cluster,
	etcdClientTLSConfig *tls.Config,
	clusterVersion string,
	bus messaging.MessageBus,
) {
	mountRouters(
		NewSubrouter(
			router.NewRoute(),
			middlewares.SimpleLogger{},
			middlewares.LimitRequest{},
		),
		routers.NewHealthRouter(actions.NewHealthController(store, cluster, etcdClientTLSConfig)),
		routers.NewVersionRouter(actions.NewVersionController(clusterVersion)),
		routers.NewTessenMetricRouter(actions.NewTessenMetricController(bus)),
	)
}

func registerAuthenticationResources(router *mux.Router, store store.Store, authenticator *authentication.Authenticator) {
	mountRouters(
		NewSubrouter(
			router.NewRoute(),
			middlewares.SimpleLogger{},
			middlewares.RefreshToken{},
			middlewares.LimitRequest{},
		),
		routers.NewAuthenticationRouter(store, authenticator),
	)
}

func mountRouters(parent *mux.Router, subRouters ...routers.Router) {
	for _, subRouter := range subRouters {
		subRouter.Mount(parent)
	}
}
