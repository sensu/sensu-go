package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/useractions"
	"github.com/sensu/sensu-go/backend/store"
)

// EventsController handles requests for /events
type EventsController struct {
	Fetcher useractions.QueryActions
}

// NewEventsController returns new EventsController
func NewEventsController(store store.EventStore) *EventsController {
	return &EventsController{
		Fetcher: &useractions.EventActions{Store: store},
	}
}

// Register the EventsController with a mux.Router.
func (c *EventsController) Register(r *mux.Router) {
	r.HandleFunc(
		"/events",
		actionHandler(c.list),
	).Methods(http.MethodGet)
	r.HandleFunc(
		"/events/{entity}",
		actionHandler(c.listByEntities),
	).Methods(http.MethodGet)
	r.HandleFunc(
		"/events/{entity}/{check}",
		actionHandler(c.find),
	).Methods(http.MethodGet)
}

func (c *EventsController) find(r *http.Request) (interface{}, error) {
	params := useractions.QueryParams(mux.Vars(r))
	fetcher := c.Fetcher.WithContext(r.Context())
	record, err := fetcher.Find(params)
	return record, err
}

func (c *EventsController) list(r *http.Request) (interface{}, error) {
	fetcher := c.Fetcher.WithContext(r.Context())
	records, err := fetcher.Query(useractions.QueryParams{})
	return records, err
}

func (c *EventsController) listByEntities(r *http.Request) (interface{}, error) {
	params := useractions.QueryParams(mux.Vars(r))
	fetcher := c.Fetcher.WithContext(r.Context())
	records, err := fetcher.Query(params)
	return records, err
}
