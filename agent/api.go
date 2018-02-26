package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
)

// APIConfig contains the API configuration
type APIConfig struct {
	Host string
	Port int
}

// newServer returns a new HTTP server
func newServer(a *Agent) *http.Server {
	router := mux.NewRouter()
	registerRoutes(a, router)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.config.API.Host, a.config.API.Port),
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return server
}

func registerRoutes(a *Agent, r *mux.Router) {
	r.HandleFunc("/events", addEvent(a)).Methods(http.MethodPost)
	r.HandleFunc("/healthz", healthz(a.conn)).Methods(http.MethodGet)
}

// healthz returns an OK status if the agent is up and connected to a backend.
// If the backend connection is closed, it returns service unavailable.
func healthz(conn transport.Transport) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if conn.Closed() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprint(w, "sensu backend unavailable")
			return
		}
		_, _ = fmt.Fprint(w, "ok")
	}
}

// addEvent accepts an event and send it to the backend over the event channel
func addEvent(a *Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var event *types.Event

		// Decode the provided event
		err := json.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Prepare the event by mutating it as required so it passes validation
		if err = prepareEvent(a, event); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		msg, err := json.Marshal(event)
		if err != nil {
			http.Error(w, fmt.Sprintf("error marshaling check result: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		a.sendMessage(transport.MessageTypeEvent, msg)

		w.WriteHeader(http.StatusCreated)
	}
}
