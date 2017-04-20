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
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// APId is the backend HTTP API.
type APId struct {
	stopping chan struct{}
	running  *atomic.Value
	wg       *sync.WaitGroup
	errChan  chan error

	Store         store.Store
	Port          int
	BackendStatus func() types.StatusMap
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

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", a.Port),
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("starting apid on address: ", server.Addr)

	go func() {
		defer a.wg.Done()
		server.ListenAndServe()
	}()

	return nil
}

// Stop httpApi.
func (a *APId) Stop() error {
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

	checksController := &controllers.ChecksController{
		Store: a.Store,
	}
	checksController.Register(r)

	entitiesController := &controllers.EntitiesController{
		Store: a.Store,
	}
	entitiesController.Register(r)

	handlersController := &controllers.HandlersController{
		Store: a.Store,
	}
	handlersController.Register(r)

	mutatorsController := &controllers.MutatorsController{
		Store: a.Store,
	}
	mutatorsController.Register(r)

	infoController := &controllers.InfoController{
		Store:  a.Store,
		Status: a.BackendStatus,
	}
	infoController.Register(r)

	healthController := &controllers.HealthController{
		Store:  a.Store,
		Status: a.BackendStatus,
	}
	healthController.Register(r)

	eventsController := &controllers.EventsController{
		Store: a.Store,
	}
	eventsController.Register(r)

	return r
}
