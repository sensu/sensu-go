package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// EventsController handles requests for /events
type EventsController struct {
	Store store.Store
}

// Register the EventsController with a mux.Router.
func (c *EventsController) Register(r *mux.Router) {
	r.HandleFunc("/events", c.events).Methods(http.MethodGet)
	r.HandleFunc("/events", c.updateEvents).Methods(http.MethodPut)
	r.HandleFunc("/events/{entity}", c.entityEvents).Methods(http.MethodGet)
	r.HandleFunc("/events/{entity}/{check}", c.entityCheckEvents).Methods(http.MethodGet)
	// TODO(greg): Handle a PUT to /events
}

func (c *EventsController) entityEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// Do we need to test that this isn't empty? We should figure that out.
	entityID := vars["entity"]

	abilities := authorization.Events.WithContext(r.Context())
	if r.Method == http.MethodGet && !abilities.CanList() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	events, err := c.Store.GetEventsByEntity(r.Context(), entityID)
	if err != nil {
		http.Error(w, "error getting events for entity", http.StatusInternalServerError)
		return
	}

	if events == nil {
		http.Error(w, "events not found", http.StatusNotFound)
		return
	}

	// Reject those resources the viewer is unauthorized to view
	rejectEvents(&events, abilities.CanRead)

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

	event, err := c.Store.GetEventByEntityCheck(r.Context(), entityID, checkID)
	if err != nil {
		http.Error(w, "error getting event for check", http.StatusInternalServerError)
		return
	}

	if event == nil {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}

	abilities := authorization.Events.WithContext(r.Context())
	if !abilities.CanRead(event) {
		authorization.UnauthorizedAccessToResource(w)
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

func (c *EventsController) events(w http.ResponseWriter, r *http.Request) {
	abilities := authorization.Events.WithContext(r.Context())
	if !abilities.CanList() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	events, err := c.Store.GetEvents(r.Context())
	if err != nil {
		http.Error(w, "error getting events", http.StatusInternalServerError)
		return
	}

	// Reject those resources the viewer is unauthorized to view
	rejectEvents(&events, abilities.CanRead)

	jsonStr, err := json.Marshal(events)
	if err != nil {
		http.Error(w, "error marshalling response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(jsonStr))
}

func (c *EventsController) updateEvents(w http.ResponseWriter, r *http.Request) {
	var event *types.Event
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(event)
	if err != nil {
		logger.Error("error decoding the body: ", err.Error())
		http.Error(w, "error decoding the body", http.StatusInternalServerError)
		return
	}

	abilities := authorization.Events.WithContext(r.Context())
	if !abilities.CanUpdate(event) {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	err = c.Store.UpdateEvent(r.Context(), event)
	if err != nil {
		logger.Error("invalid event: ", err.Error())
		http.Error(w, "invalid event", http.StatusInternalServerError)
		return
	}
}

func rejectEvents(records *[]*types.Event, predicate func(*types.Event) bool) {
	for i := 0; i < len(*records); i++ {
		if !predicate((*records)[i]) {
			*records = append((*records)[:i], (*records)[i+1:]...)
			i--
		}
	}
}
