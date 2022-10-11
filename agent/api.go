package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/lasr"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/version"
	"golang.org/x/time/rate"
)

// APIConfig contains the API configuration
type APIConfig struct {
	Host string
	Port int
}

// sensuVersion contains the API response for version
type sensuVersion struct {
	Version string `json:"version"`
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
	r.HandleFunc("/version", versionShow()).Methods(http.MethodGet)
	r.Handle("/metrics", promhttp.Handler())
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

// sensuVersion returns the version of Sensu
func versionShow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		versionJSON := sensuVersion{Version: version.Semver()}

		// Encode response
		w.Header().Set("Content-Type", "application/json")
		json, err := json.Marshal(versionJSON)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = w.Write(json)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (a *Agent) handleAPIQueue(ctx context.Context) {
	if a.config.CacheDir == os.DevNull {
		return
	}
	ch := make(chan *lasr.Message, 1)
	go func() {
		limit := a.config.EventsAPIRateLimit
		if limit == 0 {
			limit = rate.Limit(math.Inf(1))
		}
		limiter := rate.NewLimiter(limit, a.config.EventsAPIBurstLimit)
		for {
			if err := limiter.Wait(ctx); err != nil {
				// context canceled
				return
			}
			message, err := a.apiQueue.Receive(ctx)
			if err != nil {
				if ctx.Err() != nil {
					close(ch)
					return
				}
				logger.WithError(err).Error("error receiving message from queue")
				continue
			}
			ch <- message
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-ch:
			if !ok {
				return
			}
			msg := &transport.Message{
				Type:    transport.MessageTypeEvent,
				Payload: decompressMessage(message.Body),
				SendCallback: func(err error) {
					if err != nil {
						logger.WithError(err).Error("couldn't send queued message, retrying")
						_ = message.Nack(true)
					} else {
						logger.Info("queued message sent")
						_ = message.Ack()
					}
				},
			}
			a.sendMessage(msg)
		}
	}
}

// addEvent accepts an event and send it to the backend over the event channel
func addEvent(a *Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var event *corev2.Event

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

		payload, err := a.marshal(event)
		if err != nil {
			http.Error(w, fmt.Sprintf("error marshaling check result: %s", err), http.StatusInternalServerError)
			return
		}

		logEvent(event)

		if _, err := a.apiQueue.Send(compressMessage(payload)); err != nil {
			logger.WithError(err).Error("error queueing message")
			http.Error(w, "error queueing message", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
