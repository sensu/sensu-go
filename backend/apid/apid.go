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
	registerUnauthenticatedRoutes(a, router)
	registerCommonRoutes(a, router)

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

func registerUnauthenticatedRoutes(a *APId, p *mux.Router) {
	r := p.NewRoute().Subrouter()

	authenticationController := &controllers.AuthenticationController{
		Store: a.Store,
	}
	authenticationController.Register(r)
}

func registerCommonRoutes(a *APId, router *mux.Router) {
	commonRoute := router.NewRoute()
	commonRouter := commonRoute.Subrouter()

	assetsController := &controllers.AssetsController{
		Store: a.Store,
	}
	assetsController.Register(commonRouter)

	authenticationController := &controllers.AuthenticationController{
		Store: a.Store,
	}
	authenticationController.Register(commonRouter)

	checksController := &controllers.ChecksController{
		Store: a.Store,
	}
	checksController.Register(commonRouter)

	entitiesController := &controllers.EntitiesController{
		Store: a.Store,
	}
	entitiesController.Register(commonRouter)

	eventsController := &controllers.EventsController{
		Store: a.Store,
	}
	eventsController.Register(commonRouter)

	handlersController := &controllers.HandlersController{
		Store: a.Store,
	}
	handlersController.Register(commonRouter)

	healthController := &controllers.HealthController{
		Store:  a.Store,
		Status: a.BackendStatus,
	}
	healthController.Register(commonRouter)

	infoController := &controllers.InfoController{
		Store:  a.Store,
		Status: a.BackendStatus,
	}
	infoController.Register(commonRouter)

	mutatorsController := &controllers.MutatorsController{
		Store: a.Store,
	}
	mutatorsController.Register(commonRouter)

	organizationsController := &controllers.OrganizationsController{
		Store: a.Store,
	}
	organizationsController.Register(commonRouter)

	usersController := &controllers.UsersController{
		Store: a.Store,
	}
	usersController.Register(commonRouter)

	// Wrap common routes in auth & organization middleware
	commonRoute.MatcherFunc(func(r *http.Request, m *mux.RouteMatch) bool {
		// Check if the request matches any of the common routes
		if !commonRouter.Match(r, m) {
			return false
		}

		// Wrap handler in common middleware
		m.Handler = ApplyMiddleware(
			m.Handler,
			middlewares.Organization{Store: a.Store},
			middlewares.Authentication{},
			middlewares.Authorization{Store: a.Store},
			middlewares.AllowList{a.Store},
			// logging, etc.
		)

		return true
	})
}

// ApplyMiddleware apply given middleware left to right
func ApplyMiddleware(handler http.Handler, ms ...middlewares.HTTPMiddleware) http.Handler {
	var m middlewares.HTTPMiddleware

	for len(ms) > 0 {
		m, ms = ms[len(ms)-1], ms[:len(ms)-1]
		handler = m.Register(handler)
	}

	return handler
}
