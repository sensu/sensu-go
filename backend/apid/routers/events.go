package routers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// EventsRouter handles requests for /events
type EventsRouter struct {
	controller actions.EventController
}

// NewEventsRouter instantiates new events controller
func NewEventsRouter(store store.EventStore, bus messaging.MessageBus) *EventsRouter {
	return &EventsRouter{
		controller: actions.NewEventController(store, bus),
	}
}

// Mount the EventsRouter to a parent Router
func (r *EventsRouter) Mount(parent *mux.Router) {
	routes := resourceRoute{router: parent, pathPrefix: "/events"}
	routes.index(r.list)
	routes.path("{entity}", r.listByEntity).Methods(http.MethodGet)
	routes.path("{entity}/{check}", r.find).Methods(http.MethodGet)
	routes.path("{entity}/{check}", r.destroy).Methods(http.MethodDelete)
	routes.create(r.create)
}

func (r *EventsRouter) list(req *http.Request) (interface{}, error) {
	records, err := r.controller.Query(req.Context(), "", "")
	return records, err
}

func (r *EventsRouter) listByEntity(req *http.Request) (interface{}, error) {
	params := actions.QueryParams(mux.Vars(req))
	entity := url.PathEscape(params["entity"])
	records, err := r.controller.Query(req.Context(), entity, "")
	return records, err
}

func (r *EventsRouter) find(req *http.Request) (interface{}, error) {
	params := actions.QueryParams(mux.Vars(req))
	entity := url.PathEscape(params["entity"])
	check := url.PathEscape(params["check"])
	record, err := r.controller.Find(req.Context(), entity, check)
	return record, err
}

func (r *EventsRouter) destroy(req *http.Request) (interface{}, error) {
	params := actions.QueryParams(mux.Vars(req))
	entity := url.PathEscape(params["entity"])
	check := url.PathEscape(params["check"])
	return nil, r.controller.Destroy(req.Context(), entity, check)
}

func (r *EventsRouter) create(req *http.Request) (interface{}, error) {
	event := types.Event{}
	if err := unmarshalBody(req, &event); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), event)
	return event, err
}
