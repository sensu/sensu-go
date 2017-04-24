package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/store"
)

// EventsController handles requests for /events
type EventsController struct {
	Store store.Store
}

// Register the EventsController with a mux.Router.
func (c *EventsController) Register(r *mux.Router) {
	r.HandleFunc("/events/{entity}", c.entityEvents).Methods(http.MethodGet)
	r.HandleFunc("/events/{entity}/{check}", c.entityCheckEvents).Methods(http.MethodGet)
	// TODO(greg): Handle a PUT to /events
}

func (c *EventsController) entityEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityID := vars["entity"]
	// Do we need to test that this isn't empty? We should figure that out.

	events, err := c.Store.GetEventsByEntity(entityID)
	if err != nil {
		http.Error(w, "error getting events for entity", http.StatusInternalServerError)
		return
	}

	if events == nil {
		http.Error(w, "events not found", http.StatusNotFound)
		return
	}

	jsonStr, err := json.Marshal(events)
	if err != nil {
		http.Error(w, "error marshalling response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(jsonStr))
}

func (c *EventsController) entityCheckEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityID := vars["entity"]
	checkID := vars["check"]

	event, err := c.Store.GetEventByEntityCheck(entityID, checkID)
	if err != nil {
		http.Error(w, "error getting event for check", http.StatusInternalServerError)
		return
	}

	if event == nil {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}

	jsonStr, err := json.Marshal(event)
	if err != nil {
		http.Error(w, "error marshalling response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(jsonStr))
}
