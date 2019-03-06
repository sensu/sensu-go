package agent

import (
	"context"
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
	r.HandleFunc("/healthz", healthz(a.Connected)).Methods(http.MethodGet)
}

// healthz returns an OK status if the agent is up and connected to a backend.
// If the backend connection is closed, it returns service unavailable.
func healthz(connected func() bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !connected() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprint(w, "sensu backend unavailable")
			return
		}
		_, _ = fmt.Fprint(w, "ok")
	}
}

func (a *Agent) handleAPIQueue(ctx context.Context) {
	for {
		message, err := a.apiQueue.Receive(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.WithError(err).Error("error receiving message from queue")
			continue
		}
		msg := &transport.Message{
			Type:    transport.MessageTypeEvent,
			Payload: message.Body,
			SendCallback: func(err error) {
				if err != nil {
					_ = message.Nack(true)
				} else {
					_ = message.Ack()
				}
			},
		}
		a.sendMessage2(msg)
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

		payload, err := json.Marshal(event)
		if err != nil {
			http.Error(w, fmt.Sprintf("error marshaling check result: %s", err), http.StatusInternalServerError)
			return
		}

		if _, err := a.apiQueue.Send(payload); err != nil {
			logger.WithError(err).Error("error queueing message")
			http.Error(w, "error queueing message", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
