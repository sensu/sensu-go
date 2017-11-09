package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
)

// EventsRouter handles requests for /events
type EventsRouter struct {
	controller actions.EventController
}

// NewEventsRouter instantiates new events controller
func NewEventsRouter(store store.EventStore) *EventsRouter {
	return &EventsRouter{
		controller: actions.NewEventController(store),
	}
}

// Mount the EventsRouter to a parent Router
func (r *EventsRouter) Mount(parent *mux.Router) {
	routes := resourceRoute{router: parent, pathPrefix: "/events"}
	routes.index(r.list)
	routes.path("{entity}", r.listByEntity).Methods(http.MethodGet)
	routes.path("{entity}/{check}", r.find).Methods(http.MethodGet)
}

func (r *EventsRouter) list(req *http.Request) (interface{}, error) {
	records, err := r.controller.Query(req.Context(), actions.QueryParams{})
	return records, err
}

func (r *EventsRouter) listByEntity(req *http.Request) (interface{}, error) {
	params := actions.QueryParams(mux.Vars(req))
	records, err := r.controller.Query(req.Context(), params)
	return records, err
}

func (r *EventsRouter) find(req *http.Request) (interface{}, error) {
	params := actions.QueryParams(mux.Vars(req))
	record, err := r.controller.Find(req.Context(), params)
	return record, err
}
