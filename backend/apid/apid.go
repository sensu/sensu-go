package apid

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/controllers"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
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

	a.stopping = make(chan struct{}, 1)
	a.running = &atomic.Value{}
	a.wg = &sync.WaitGroup{}

	a.errChan = make(chan error, 1)

	router := mux.NewRouter()
	registerAuthenticationResources(router, a.Store)
	registerRestrictedResources(router, a.Store, a.BackendStatus)

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

func registerAuthenticationResources(router *mux.Router, store store.Store) {
	authRouter := NewSubrouter(
		router.NewRoute(),
		middlewares.SimpleLogger{},
		middlewares.RefreshToken{},
	)

	authenticationController := controllers.AuthenticationController{Store: store}
	authenticationController.Register(authRouter)
}

func registerRestrictedResources(
	router *mux.Router,
	store store.Store,
	bStatus func() types.StatusMap,
) {
	commonRouter := NewSubrouter(
		router.NewRoute(),
		middlewares.SimpleLogger{},
		middlewares.Environment{Store: store},
		middlewares.Authentication{},
		middlewares.AllowList{Store: store},
		middlewares.Authorization{Store: store},
	)

	assetsController := &controllers.AssetsController{
		Store: store,
	}
	assetsController.Register(commonRouter)

	authenticationController := &controllers.AuthenticationController{
		Store: store,
	}
	authenticationController.Register(commonRouter)

	checksController := &controllers.ChecksController{
		Store: store,
	}
	checksController.Register(commonRouter)

	entitiesController := &controllers.EntitiesController{
		Store: store,
	}
	entitiesController.Register(commonRouter)

	environmentsConroller := &controllers.EnvironmentsController{
		Store: store,
	}
	environmentsConroller.Register(commonRouter)

	eventsController := controllers.NewEventsController(store)
	eventsController.Register(commonRouter)

	handlersController := &controllers.HandlersController{
		Store: store,
	}
	handlersController.Register(commonRouter)

	healthController := &controllers.HealthController{
		Store:  store,
		Status: bStatus,
	}
	healthController.Register(commonRouter)

	infoController := &controllers.InfoController{
		Store:  store,
		Status: bStatus,
	}
	infoController.Register(commonRouter)

	mutatorsController := &controllers.MutatorsController{
		Store: store,
	}
	mutatorsController.Register(commonRouter)

	organizationsController := &controllers.OrganizationsController{
		Store: store,
	}
	organizationsController.Register(commonRouter)

	rolesController := &controllers.RolesController{
		Store: store,
	}
	rolesController.Register(commonRouter)

	usersController := &controllers.UsersController{
		Store: store,
	}
	usersController.Register(commonRouter)

	graphqlController := &controllers.GraphController{Store: store}
	graphqlController.Register(commonRouter)
}
