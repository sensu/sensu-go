package apid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
	"github.com/sensu/sensu-go/backend/apid/routers"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// APId is the backend HTTP API.
type APId struct {
	// Host is the host APId is running on.
	Host string

	// Port is the port APId is running on.
	Port int

	stopping      chan struct{}
	running       *atomic.Value
	wg            *sync.WaitGroup
	errChan       chan error
	httpServer    *http.Server
	bus           messaging.MessageBus
	backendStatus func() types.StatusMap
	store         store.Store
	queueGetter   types.QueueGetter
	tls           *types.TLSOptions
}

// Option is a functional option.
type Option func(*APId) error

// Config configures APId.
type Config struct {
	Host          string
	Port          int
	Bus           messaging.MessageBus
	Store         store.Store
	QueueGetter   types.QueueGetter
	TLS           *types.TLSOptions
	BackendStatus func() types.StatusMap
}

// New creates a new APId.
func New(c Config, opts ...Option) (*APId, error) {
	a := &APId{
		Host:          c.Host,
		Port:          c.Port,
		store:         c.Store,
		queueGetter:   c.QueueGetter,
		tls:           c.TLS,
		backendStatus: c.BackendStatus,
		bus:           c.Bus,
		stopping:      make(chan struct{}, 1),
		running:       &atomic.Value{},
		wg:            &sync.WaitGroup{},
		errChan:       make(chan error, 1),
	}

	router := mux.NewRouter().UseEncodedPath()
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	registerUnauthenticatedResources(router, a.backendStatus)
	registerAuthenticationResources(router, a.store)
	registerRestrictedResources(router, a.store, a.queueGetter, a.bus)

	a.httpServer = &http.Server{
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
	logger.Info("starting apid on address: ", a.httpServer.Addr)
	a.wg.Add(1)

	go func() {
		defer a.wg.Done()
		var err error
		if a.tls != nil {
			err = a.httpServer.ListenAndServeTLS(a.tls.CertFile, a.tls.KeyFile)
		} else {
			err = a.httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			a.errChan <- fmt.Errorf("failed to start http/https server %s", err.Error())
		}
	}()

	return nil
}

// Stop httpApi.
func (a *APId) Stop() error {
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

// Status returns an error if httpApi is unhealthy.
func (a *APId) Status() error {
	return nil
}

// Err returns a channel to listen for terminal errors on.
func (a *APId) Err() <-chan error {
	return a.errChan
}

func registerUnauthenticatedResources(
	router *mux.Router,
	bStatus func() types.StatusMap,
) {
	mountRouters(
		NewSubrouter(
			router.NewRoute(),
			middlewares.SimpleLogger{},
			middlewares.LimitRequest{},
		),
		routers.NewStatusRouter(bStatus),
	)
}

func registerAuthenticationResources(router *mux.Router, store store.Store) {
	mountRouters(
		NewSubrouter(
			router.NewRoute(),
			middlewares.SimpleLogger{},
			middlewares.RefreshToken{},
			middlewares.LimitRequest{},
		),
		routers.NewAuthenticationRouter(store),
	)
}

func registerRestrictedResources(router *mux.Router, store store.Store, getter types.QueueGetter, bus messaging.MessageBus) {
	mountRouters(
		NewSubrouter(
			router.NewRoute(),
			middlewares.SimpleLogger{},
			middlewares.Environment{Store: store},
			middlewares.Authentication{},
			middlewares.AllowList{Store: store},
			middlewares.Authorization{Store: store},
			middlewares.LimitRequest{},
		),
		routers.NewAssetRouter(store),
		routers.NewChecksRouter(store, getter),
		routers.NewEntitiesRouter(store),
		routers.NewEnvironmentsRouter(store),
		routers.NewEventFiltersRouter(store),
		routers.NewEventsRouter(store, bus),
		routers.NewGraphQLRouter(store, bus, getter),
		routers.NewHandlersRouter(store),
		routers.NewHooksRouter(store),
		routers.NewMutatorsRouter(store),
		routers.NewOrganizationsRouter(store),
		routers.NewRolesRouter(store),
		routers.NewSilencedRouter(store),
		routers.NewUsersRouter(store),
		routers.NewExtensionsRouter(store),
	)
}

func mountRouters(parent *mux.Router, subRouters ...routers.Router) {
	for _, subRouter := range subRouters {
		subRouter.Mount(parent)
	}
}
