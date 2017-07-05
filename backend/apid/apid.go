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
	"github.com/sensu/sensu-go/backend/authentication"
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

	Authentication authentication.Provider
	BackendStatus  func() types.StatusMap
	Host           string
	Port           int
	Store          store.Store
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

	router := httpRouter(a)
	routerStack := authentication.Middleware(a.Authentication, router)
	routerStack = middlewares.Organization(routerStack, a.Store)

	a.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.Host, a.Port),
		Handler:      routerStack,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("starting apid on address: ", a.httpServer.Addr)
	a.wg.Add(1)

	go func() {
		defer a.wg.Done()
		if err := a.httpServer.ListenAndServe(); err != nil {
			// TODO (JK): need a way to handle closing things like errChan, etc.
			// in cases where there's a failure to start the daemon
			logger.Errorf("failed to start http server: %s", err.Error())
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

func httpRouter(a *APId) *mux.Router {
	r := mux.NewRouter()

	assetsController := &controllers.AssetsController{
		Store: a.Store,
	}
	assetsController.Register(r)

	authenticationController := &controllers.AuthenticationController{
		Provider: a.Authentication,
	}
	authenticationController.Register(r)

	checksController := &controllers.ChecksController{
		Store: a.Store,
	}
	checksController.Register(r)

	entitiesController := &controllers.EntitiesController{
		Store: a.Store,
	}
	entitiesController.Register(r)

	eventsController := &controllers.EventsController{
		Store: a.Store,
	}
	eventsController.Register(r)

	handlersController := &controllers.HandlersController{
		Store: a.Store,
	}
	handlersController.Register(r)

	healthController := &controllers.HealthController{
		Store:  a.Store,
		Status: a.BackendStatus,
	}
	healthController.Register(r)

	infoController := &controllers.InfoController{
		Store:  a.Store,
		Status: a.BackendStatus,
	}
	infoController.Register(r)

	mutatorsController := &controllers.MutatorsController{
		Store: a.Store,
	}
	mutatorsController.Register(r)

	organizationsController := &controllers.OrganizationsController{
		Store: a.Store,
	}
	organizationsController.Register(r)

	usersController := &controllers.UsersController{
		Provider: a.Authentication,
		Store:    a.Store,
	}
	usersController.Register(r)

	return r
}
