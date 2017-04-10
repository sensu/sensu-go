package api

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/api/controllers"
	"github.com/sensu/sensu-go/backend/store"
)

// HttpApi is the backend HTTP API.
type HttpApi struct {
	stopping  chan struct{}
	running   *atomic.Value
	wg        *sync.WaitGroup
	errChan   chan error
	eventChan chan []byte

	Store store.Store
	Port  int
}

// Start httpApi.
func (a *HttpApi) Start() error {
	if a.Store == nil {
		return errors.New("no store found")
	}

	a.stopping = make(chan struct{}, 1)
	a.running = &atomic.Value{}
	a.wg = &sync.WaitGroup{}

	a.errChan = make(chan error, 1)
	a.eventChan = make(chan []byte, 100)

	router := httpRouter(a)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", a.Port),
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		defer a.wg.Done()
		server.ListenAndServe()
	}()

	return nil
}

// Stop httpApi.
func (a *HttpApi) Stop() error {
	a.running.Store(false)
	close(a.stopping)
	a.wg.Wait()
	close(a.errChan)
	close(a.eventChan)

	return nil
}

// Status returns an error if httpApi is unhealthy.
func (a *HttpApi) Status() error {
	return nil
}

// Err returns a channel to listen for terminal errors on.
func (a *HttpApi) Err() <-chan error {
	return a.errChan
}

// StatusMap is a map of backend component names to their current status info.
//type StatusMap map[string]bool

// Healthy returns true if the StatsMap shows all healthy indicators; false
// otherwise.
/*func (s StatusMap) Healthy() bool {
	for _, v := range s {
		if !v {
			return false
		}
	}
	return true
}*/

// InfoHandler handles GET requests to the /info endpoint.
/*func (a *HttpApi) InfoHandler(w http.ResponseWriter, r *http.Request) {
	sb, err := json.Marshal(a.Status())
	if err != nil {
		log.Println("error marshaling status: ", err.Error())
		http.Error(w, "Error getting server status.", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(sb))
}*/

// HealthHandler handles GET requests to the /health endpoint.
/*func (a *HttpApi) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if !a.Status().Healthy() {
		http.Error(w, "", http.StatusServiceUnavailable)
		return
	}
	// implicitly returns 200
}*/

func httpRouter(api *HttpApi) *mux.Router {
	r := mux.NewRouter()

	checksController := &controllers.ChecksController{
		Store: api.Store,
	}
	checksController.Register(r)

	entitiesController := &controllers.EntitiesController{
		Store: api.Store,
	}
	entitiesController.Register(r)

	handlersController := &controllers.HandlersController{
		Store: api.Store,
	}
	handlersController.Register(r)

	mutatorsController := &controllers.MutatorsController{
		Store: api.Store,
	}
	mutatorsController.Register(r)

	//r.HandleFunc("/info", api.InfoHandler).Methods(http.MethodGet)
	//r.HandleFunc("/health", api.HealthHandler).Methods(http.MethodGet)

	return r
}
