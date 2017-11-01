package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/useractions"
	"github.com/sensu/sensu-go/backend/store"
)

// EventsController handles requests for /events
type EventsController struct {
	Store store.EventStore
}

// Register the EventsController with a mux.Router.
func (c *EventsController) Register(r *mux.Router) { // TODO: Rename to 'Mount'?
	routes := resourceRoute{router: r, pathPrefix: "/events"}
	routes.index(c.list)
	routes.path("{entity}", c.listByEntity).Methods(http.MethodGet)
	routes.path("{entity}/{check}", c.find).Methods(http.MethodGet)
}

func (c *EventsController) find(r *http.Request) (interface{}, error) {
	fetcher := useractions.NewEventActions(r.Context(), c.Store)
	params := useractions.QueryParams(mux.Vars(r))
	record, err := fetcher.Find(params)
	return record, err
}

func (c *EventsController) list(r *http.Request) (interface{}, error) {
	fetcher := useractions.NewEventActions(r.Context(), c.Store)
	records, err := fetcher.Query(useractions.QueryParams{})
	return records, err
}

func (c *EventsController) listByEntity(r *http.Request) (interface{}, error) {
	fetcher := useractions.NewEventActions(r.Context(), c.Store)
	params := useractions.QueryParams(mux.Vars(r))
	records, err := fetcher.Query(params)
	return records, err
}
