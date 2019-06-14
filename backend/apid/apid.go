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

	// prepare TLS configs (both server and client)
	var tlsServerConfig, tlsClientConfig *tls.Config
	var err error
	if c.TLS != nil {
		tlsServerConfig, err = c.TLS.ToServerTLSConfig()
		if err != nil {
			return nil, err
		}

		// TODO(gbolo): since the backend does not yet support mutual TLS
		// this client does not need a cert and key. Once we do support
		// mutual TLS we need new options for client cert and key
		cTLSOptions := types.TLSOptions{
			TrustedCAFile: c.TLS.TrustedCAFile,
			// TODO(palourde): We should avoid using the loopback interface
			InsecureSkipVerify: true,
		}
		tlsClientConfig, err = cTLSOptions.ToClientTLSConfig()
		if err != nil {
			return nil, err
		}
	}

	router := mux.NewRouter().UseEncodedPath()
	router.NotFoundHandler = middlewares.SimpleLogger{}.Then(http.HandlerFunc(notFoundHandler))
	router.Handle("/metrics", promhttp.Handler())
	registerUnauthenticatedResources(router, a.store, a.cluster, a.etcdClientTLSConfig, a.clusterVersion, a.bus)
	a.registerGraphQLService(router, c.URL, tlsClientConfig)
	registerAuthenticationResources(router, a.store, a.Authenticator)
	a.registerRestrictedResources(router)

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
		routers.NewClusterIDRouter(actions.NewClusterIDController(store)),
	)
}

func (a *APId) registerGraphQLService(router *mux.Router, url string, tls *tls.Config) {
	a.GraphQLSubrouter = NewSubrouter(
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
		middlewares.AllowList{Store: a.store, IgnoreMissingClaims: true},
	)
	mountRouters(
		a.GraphQLSubrouter,
		routers.NewGraphQLRouter(url, tls, a.store),
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

func (a *APId) registerRestrictedResources(router *mux.Router) {
	a.CoreSubrouter = NewSubrouter(
		router.NewRoute().
			PathPrefix("/api/{group:core}/{version:v2}/"),
		middlewares.SimpleLogger{},
		middlewares.Namespace{},
		middlewares.Authentication{},
		middlewares.AllowList{Store: a.store},
		middlewares.AuthorizationAttributes{},
		middlewares.Authorization{Authorizer: &rbac.Authorizer{Store: a.store}},
		middlewares.LimitRequest{},
		middlewares.Pagination{},
	)
	mountRouters(
		a.CoreSubrouter,
		routers.NewAssetRouter(a.store),
		routers.NewChecksRouter(a.store, a.queueGetter),
		routers.NewClusterRolesRouter(a.store),
		routers.NewClusterRoleBindingsRouter(a.store),
		routers.NewClusterRouter(actions.NewClusterController(a.cluster)),
		routers.NewEntitiesRouter(a.store),
		routers.NewEventFiltersRouter(a.store),
		routers.NewEventsRouter(a.eventStore, a.bus),
		routers.NewExtensionsRouter(a.store),
		routers.NewHandlersRouter(a.store),
		routers.NewHooksRouter(a.store),
		routers.NewMutatorsRouter(a.store),
		routers.NewNamespacesRouter(a.store),
		routers.NewRolesRouter(a.store),
		routers.NewRoleBindingsRouter(a.store),
		routers.NewSilencedRouter(a.store),
		routers.NewTessenRouter(actions.NewTessenController(a.store, a.bus)),
		routers.NewUsersRouter(a.store),
	)
}

func mountRouters(parent *mux.Router, subRouters ...routers.Router) {
	for _, subRouter := range subRouters {
		subRouter.Mount(parent)
	}
}
