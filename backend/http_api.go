package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// StatusMap is a map of backend component names to their current status info.
type StatusMap map[string]bool

// Healthy returns true if the StatsMap shows all healthy indicators; false
// otherwise.
func (s StatusMap) Healthy() bool {
	for _, v := range s {
		if !v {
			return false
		}
	}
	return true
}

// API is the backend HTTP API.
type API struct {
	Status func() types.StatusMap
	Store  store.Store
}

// InfoHandler handles GET requests to the /info endpoint.
func (a *API) InfoHandler(w http.ResponseWriter, r *http.Request) {
	sb, err := json.Marshal(a.Status())
	if err != nil {
		log.Println("error marshaling status: ", err.Error())
		http.Error(w, "Error getting server status.", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(sb))
}

// HealthHandler handles GET requests to the /health endpoint.
func (a *API) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if !a.Status().Healthy() {
		http.Error(w, "", http.StatusServiceUnavailable)
		return
	}
	// implicitly returns 200
}

// EntitiesHandler handles GET requests to the /entities endpoint.
func (a *API) EntitiesHandler(w http.ResponseWriter, r *http.Request) {
	es, err := a.Store.GetEntities()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	esb, err := json.Marshal(es)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(esb))
}

// EntityHandler handles requests to /entities/{id}.
func (a *API) EntityHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	entity, err := a.Store.GetEntityByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entity == nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	eb, err := json.Marshal(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(eb))
}

// CheckHandler handles requests to /checks/:name
func (a *API) CheckHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	method := r.Method

	var (
		check *types.Check
		err   error
	)

	if method == http.MethodGet || method == http.MethodDelete {
		check, err = a.Store.GetCheckByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		if check == nil {
			http.NotFound(w, r)
			return
		}

		checkBytes, err := json.Marshal(check)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, string(checkBytes))
	case http.MethodPut, http.MethodPost:
		newCheck := &types.Check{}
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal(bodyBytes, newCheck)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = newCheck.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = a.Store.UpdateCheck(newCheck)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	case http.MethodDelete:
		if check == nil {
			http.NotFound(w, r)
		}

		err := a.Store.DeleteCheckByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
}

// ChecksHandler handles requests to /checks
func (a *API) ChecksHandler(w http.ResponseWriter, r *http.Request) {
	checks, err := a.Store.GetChecks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	checksBytes, err := json.Marshal(checks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprintf(w, string(checksBytes))
}

// HandlerHandler handles requests to /handlers/:name
func (a *API) HandlerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	method := r.Method

	var (
		handler *types.Handler
		err     error
	)

	if method == http.MethodGet || method == http.MethodDelete {
		handler, err = a.Store.GetHandlerByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if handler == nil {
			http.NotFound(w, r)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		handlerBytes, err := json.Marshal(handler)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, string(handlerBytes))
	case http.MethodPut, http.MethodPost:
		newHandler := &types.Handler{}
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal(bodyBytes, newHandler)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = newHandler.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = a.Store.UpdateHandler(newHandler)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
		err := a.Store.DeleteHandlerByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// HandlersHandler handles requests to /handlers
func (a *API) HandlersHandler(w http.ResponseWriter, r *http.Request) {
	handlers, err := a.Store.GetHandlers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	handlersBytes, err := json.Marshal(handlers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprintf(w, string(handlersBytes))
}

// MutatorsHandler handles requests to /mutators
func (a *API) MutatorsHandler(w http.ResponseWriter, r *http.Request) {
	mutators, err := a.Store.GetMutators()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mutatorsBytes, err := json.Marshal(mutators)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprintf(w, string(mutatorsBytes))
}

// MutatorHandler handles requests to /mutators/:name
func (a *API) MutatorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	method := r.Method

	var (
		mutator *types.Mutator
		err     error
	)

	if method == http.MethodGet || method == http.MethodDelete {
		mutator, err = a.Store.GetMutatorByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if mutator == nil {
			http.NotFound(w, r)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		mutatorBytes, err := json.Marshal(mutator)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, string(mutatorBytes))
	case http.MethodPut, http.MethodPost:
		newMutator := &types.Mutator{}
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal(bodyBytes, newMutator)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = newMutator.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = a.Store.UpdateMutator(newMutator)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
		err := a.Store.DeleteMutatorByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func httpServer(b *Backend) (*http.Server, error) {
	store, err := b.etcd.NewStore()
	if err != nil {
		return nil, err
	}

	api := &API{
		Status: b.Status,
		Store:  store,
	}

	router := httpRouter(api)

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", b.Config.APIPort),
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}, nil
}

func httpRouter(api *API) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/info", api.InfoHandler).Methods(http.MethodGet)
	r.HandleFunc("/health", api.HealthHandler).Methods(http.MethodGet)
	r.HandleFunc("/entities", api.EntitiesHandler).Methods(http.MethodGet)
	r.HandleFunc("/entities/{id}", api.EntityHandler).Methods(http.MethodGet)
	r.HandleFunc("/checks", api.ChecksHandler).Methods(http.MethodGet)
	r.HandleFunc("/checks/{name}", api.CheckHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete)
	r.HandleFunc("/handlers", api.HandlersHandler).Methods(http.MethodGet)
	r.HandleFunc("/handlers/{name}", api.HandlerHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete)
	r.HandleFunc("/mutators", api.MutatorsHandler).Methods(http.MethodGet)
	r.HandleFunc("/mutators/{name}", api.MutatorHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete)

	return r
}
