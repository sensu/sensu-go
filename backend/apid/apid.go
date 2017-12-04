package apid

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
	"github.com/sensu/sensu-go/backend/apid/routers"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// APId is the backend HTTP API.
type APId struct {
	stopping   chan struct{}
	running    *atomic.Value
	wg         *sync.WaitGroup
	errChan    chan error
	httpServer *http.Server
	MessageBus messaging.MessageBus

	BackendStatus func() types.StatusMap
	Host          string
	Port          int
	Store         store.Store
	TLS           *types.TLSOptions
}

// Start Apid.
func (a *APId) Start() error {
	if a.Store == nil {
		return errors.New("no store found")
	}

	if a.MessageBus == nil {
		return errors.New("no message bus found")
	}

	a.stopping = make(chan struct{}, 1)
	a.running = &atomic.Value{}
	a.wg = &sync.WaitGroup{}

	a.errChan = make(chan error, 1)

	router := mux.NewRouter()
	registerUnauthenticatedResources(router, a.BackendStatus)
	registerAuthenticationResources(router, a.Store)
	registerRestrictedResources(router, a.Store, a.MessageBus)

	a.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.Host, a.Port),
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("starting apid on address: ", a.httpServer.Addr)
	a.wg.Add(1)

	go func() {
		defer a.wg.Done()
		var err error
		if a.TLS != nil {
			err = a.httpServer.ListenAndServeTLS(a.TLS.CertFile, a.TLS.KeyFile)
		} else {
			err = a.httpServer.ListenAndServe()
		}
		// TODO (JK): need a way to handle closing things like errChan, etc.
		// in cases where there's a failure to start the daemon
		if err != nil && err != http.ErrServerClosed {
			a.errChan <- fmt.Errorf("failed to start http/https server %s", err.Error())
		}
	}()

	return nil
}

// Stop httpApi.
func (a *APId) Stop() error {
	if err := a.httpServer.Shutdown(nil); err != nil {
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

func registerRestrictedResources(router *mux.Router, store store.Store, bus messaging.MessageBus) {
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
		routers.NewChecksRouter(store),
		routers.NewEntitiesRouter(store),
		routers.NewEnvironmentsRouter(store),
		routers.NewEventFiltersRouter(store),
		routers.NewEventsRouter(store, bus),
		routers.NewGraphQLRouter(store),
		routers.NewHandlersRouter(store),
		routers.NewMutatorsRouter(store),
		routers.NewOrganizationsRouter(store),
		routers.NewRolesRouter(store),
		routers.NewSilencedRouter(store),
		routers.NewUsersRouter(store),
	)
}

func mountRouters(parent *mux.Router, subRouters ...routers.Router) {
	for _, subRouter := range subRouters {
		subRouter.Mount(parent)
	}
}
