package agent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
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
}

// addEvent accepts an event and send it to the backend over the event channel
func addEvent(a *Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var event *types.Event

		// Decode the provided event
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal(bodyBytes, &event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Make sure the event is valid
		if err := validateEvent(a, event); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		msg, err := json.Marshal(event)
		if err != nil {
			http.Error(w, fmt.Sprintf("error marshaling check result: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		a.sendMessage(types.EventType, msg)

		w.WriteHeader(http.StatusCreated)
		return
	}
}
