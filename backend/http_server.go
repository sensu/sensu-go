package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/store"
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
	Status func() StatusMap
	Store  store.Store
}

// InfoHandler handles GET requests to the /info endpoint.
func (a *API) InfoHandler(w http.ResponseWriter, r *http.Request) {
	sb, err := json.Marshal(a.Status())
	if err != nil {
		log.Println("error marshaling status: ", err.Error())
		http.Error(w, "Error getting server status.", http.StatusInternalServerError)
	}
	fmt.Fprint(w, sb)
}

// HealthHandler handles GET requests to the /health endpoint.
func (a *API) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if !a.Status().Healthy() {
		http.Error(w, "", http.StatusServiceUnavailable)
	}
	// implicitly returns 200
}

// EntitiesHandler handles GET requests to the /entities endpoint.
func (a *API) EntitiesHandler(w http.ResponseWriter, r *http.Request) {
	es, err := a.Store.GetEntities()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	esb, err := json.Marshal(es)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprint(w, esb)
}

// EntityHandler handles requests to /entities/{id}.
func (a *API) EntityHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	entity, err := a.Store.GetEntityByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if entity == nil {
		http.Error(w, "", http.StatusNotFound)
	}

	eb, err := json.Marshal(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprint(w, eb)
}

func httpServer(b *Backend) *http.Server {
	api := &API{
		Status: b.Status,
		Store:  b.store,
	}

	r := mux.NewRouter()

	r.HandleFunc("/info", api.InfoHandler).Methods("GET")
	r.HandleFunc("/health", api.HealthHandler).Methods("GET")
	r.HandleFunc("/entities", api.EntitiesHandler).Methods("GET")
	r.HandleFunc("/entities/{id}", api.EntityHandler).Methods("GET")

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", b.Config.APIPort),
		Handler:      r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}
