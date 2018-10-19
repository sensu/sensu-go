package apid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
	"github.com/sensu/sensu-go/backend/apid/routers"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/version"
)

// APId is the backend HTTP API.
type APId struct {
	// Host is the host APId is running on.
	Host string

	// Port is the port APId is running on.
	Port int

	stopping    chan struct{}
	running     *atomic.Value
	wg          *sync.WaitGroup
	errChan     chan error
	HTTPServer  *http.Server
	bus         messaging.MessageBus
	store       store.Store
	queueGetter types.QueueGetter
	tls         *types.TLSOptions
	cluster     clientv3.Cluster
}

// Option is a functional option.
type Option func(*APId) error

// Config configures APId.
type Config struct {
	Host        string
	Port        int
	Bus         messaging.MessageBus
	Store       store.Store
	QueueGetter types.QueueGetter
	TLS         *types.TLSOptions
	Cluster     clientv3.Cluster
}

// New creates a new APId.
func New(c Config, opts ...Option) (*APId, error) {
	a := &APId{
		Host:        c.Host,
		Port:        c.Port,
		store:       c.Store,
		queueGetter: c.QueueGetter,
		tls:         c.TLS,
		bus:         c.Bus,
		stopping:    make(chan struct{}, 1),
		running:     &atomic.Value{},
		wg:          &sync.WaitGroup{},
		errChan:     make(chan error, 1),
		cluster:     c.Cluster,
	}

	router := mux.NewRouter().UseEncodedPath()
	router.NotFoundHandler = middlewares.SimpleLogger{}.Then(http.HandlerFunc(notFoundHandler))
	registerUnauthenticatedResources(router, a.store)
	registerAuthenticationResources(router, a.store)
	registerRestrictedResources(router, a.store, a.queueGetter, a.bus, a.cluster)

	a.HTTPServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.Host, a.Port),
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
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
		"error": "not found", "code": actions.NotFound,
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
			err = a.HTTPServer.ListenAndServeTLS(a.tls.CertFile, a.tls.KeyFile)
		} else {
			err = a.HTTPServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			a.errChan <- fmt.Errorf("failed to start http/https server %s", err.Error())
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
) {
	mountRouters(
		NewSubrouter(
			router.NewRoute(),
			middlewares.SimpleLogger{},
			middlewares.LimitRequest{},
			middlewares.Edition{Name: version.Edition},
		),
		routers.NewHealthRouter(actions.NewHealthController(store)),
	)
}

func registerAuthenticationResources(router *mux.Router, store store.Store) {
	mountRouters(
		NewSubrouter(
			router.NewRoute(),
			middlewares.SimpleLogger{},
			middlewares.RefreshToken{},
			middlewares.LimitRequest{},
			middlewares.Edition{Name: version.Edition},
		),
		routers.NewAuthenticationRouter(store),
	)
}

func registerRestrictedResources(router *mux.Router, store store.Store, getter types.QueueGetter, bus messaging.MessageBus, cluster clientv3.Cluster) {
	mountRouters(
		NewSubrouter(
			router.NewRoute(),
			middlewares.SimpleLogger{},
			middlewares.Namespace{Store: store},
			middlewares.Authentication{},
			middlewares.AllowList{Store: store},
			middlewares.Authorization{Store: store},
			middlewares.LimitRequest{},
			middlewares.Edition{Name: version.Edition},
			middlewares.RequestInfo{},
		),
		routers.NewAssetRouter(store),
		routers.NewChecksRouter(actions.NewCheckController(store, getter)),
		routers.NewEntitiesRouter(store),
		routers.NewEventFiltersRouter(store),
		routers.NewEventsRouter(store, bus),
		routers.NewGraphQLRouter(store, bus, getter),
		routers.NewHandlersRouter(store),
		routers.NewHooksRouter(store),
		routers.NewMutatorsRouter(store),
		routers.NewNamespacesRouter(actions.NewNamespacesController(store)),
		routers.NewRolesRouter(store),
		routers.NewSilencedRouter(store),
		routers.NewUsersRouter(store),
		routers.NewExtensionsRouter(store),
		routers.NewClusterRouter(actions.NewClusterController(cluster)),
	)
}

func mountRouters(parent *mux.Router, subRouters ...routers.Router) {
	for _, subRouter := range subRouters {
		subRouter.Mount(parent)
	}
}
