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

	"github.com/sensu/sensu-go/backend/authentication"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
	"github.com/sensu/sensu-go/backend/apid/routers"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// APId is the backend HTTP API.
type APId struct {
	Authenticator *authentication.Authenticator
	HTTPServer    *http.Server

	stopping            chan struct{}
	running             *atomic.Value
	wg                  *sync.WaitGroup
	errChan             chan error
	bus                 messaging.MessageBus
	store               store.Store
	queueGetter         types.QueueGetter
	tls                 *types.TLSOptions
	cluster             clientv3.Cluster
	etcdClientTLSConfig *tls.Config
}

// Option is a functional option.
type Option func(*APId) error

// Config configures APId.
type Config struct {
	ListenAddress       string
	URL                 string
	Bus                 messaging.MessageBus
	Store               store.Store
	QueueGetter         types.QueueGetter
	TLS                 *types.TLSOptions
	Cluster             clientv3.Cluster
	EtcdClientTLSConfig *tls.Config
	Authenticator       *authentication.Authenticator
}

// New creates a new APId.
func New(c Config, opts ...Option) (*APId, error) {
	a := &APId{
		store:               c.Store,
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
	registerUnauthenticatedResources(router, a.store, a.cluster, a.etcdClientTLSConfig)
	registerGraphQLService(router, a.store, c.URL, tlsClientConfig)
	registerAuthenticationResources(router, a.store, a.Authenticator)
	registerRestrictedResources(router, a.store, a.queueGetter, a.bus, a.cluster)

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
) {
	mountRouters(
		NewSubrouter(
			router.NewRoute(),
			middlewares.SimpleLogger{},
			middlewares.LimitRequest{},
		),
		routers.NewHealthRouter(actions.NewHealthController(store, cluster, etcdClientTLSConfig)),
	)
}

func registerGraphQLService(router *mux.Router, store store.Store, url string, tls *tls.Config) {
	mountRouters(
		NewSubrouter(
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
			middlewares.AllowList{Store: store, IgnoreMissingClaims: true},
		),
		routers.NewGraphQLRouter(url, tls, store),
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

func registerRestrictedResources(router *mux.Router, store store.Store, getter types.QueueGetter, bus messaging.MessageBus, cluster clientv3.Cluster) {
	mountRouters(
		NewSubrouter(
			router.NewRoute().
				PathPrefix("/api/{group:core}/{version:v2}/"),
			middlewares.SimpleLogger{},
			middlewares.Namespace{},
			middlewares.Authentication{},
			middlewares.AllowList{Store: store},
			middlewares.AuthorizationAttributes{},
			middlewares.Authorization{Authorizer: &rbac.Authorizer{Store: store}},
			middlewares.LimitRequest{},
		),
		routers.NewAssetRouter(store),
		routers.NewChecksRouter(actions.NewCheckController(store, getter)),
		routers.NewClusterRolesRouter(store),
		routers.NewClusterRoleBindingsRouter(store),
		routers.NewClusterRouter(actions.NewClusterController(cluster)),
		routers.NewEntitiesRouter(store),
		routers.NewEventFiltersRouter(store),
		routers.NewEventsRouter(store, bus),
		routers.NewExtensionsRouter(store),
		routers.NewHandlersRouter(store),
		routers.NewHooksRouter(store),
		routers.NewMutatorsRouter(store),
		routers.NewNamespacesRouter(actions.NewNamespacesController(store)),
		routers.NewRolesRouter(store),
		routers.NewRoleBindingsRouter(store),
		routers.NewSilencedRouter(store),
		routers.NewUsersRouter(store),
	)
	mountRouters(
		NewSubrouter(
			router.NewRoute(),
			middlewares.SimpleLogger{},
			middlewares.Authentication{},
			middlewares.AllowList{Store: store},
			middlewares.AuthorizationAttributes{},
			middlewares.Authorization{Authorizer: &rbac.Authorizer{Store: store}},
			middlewares.LimitRequest{},
		),
		routers.NewTessenRouter(actions.NewTessenController(store)),
	)
}

func mountRouters(parent *mux.Router, subRouters ...routers.Router) {
	for _, subRouter := range subRouters {
		subRouter.Mount(parent)
	}
}
