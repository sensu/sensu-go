package routers

import (
	"net/http"
	"net/url"
	"path"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
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
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:events}",
	}

	routes.Post(r.create)
	routes.List(r.controller.List, corev2.EventFields)
	routes.ListAllNamespaces(r.controller.List, "/{resource:events}", corev2.EventFields)
	routes.Path("{entity}/{check}", r.find).Methods(http.MethodGet)
	routes.Path("{entity}/{check}", r.destroy).Methods(http.MethodDelete)
	routes.Path("{entity}/{check}", r.createOrReplace).Methods(http.MethodPut)

	// Additionaly allow a subcollection to be specified when listing events,
	// which correspond to the entity name here
	parent.HandleFunc(path.Join(routes.PathPrefix, "{subcollection}"),
		listerHandler(r.controller.List, corev2.EventFields)).Methods(http.MethodGet)
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
	if err := UnmarshalBody(req, &event); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), event)
	return event, err
}

func (r *EventsRouter) createOrReplace(req *http.Request) (interface{}, error) {
	event := types.Event{}
	if err := UnmarshalBody(req, &event); err != nil {
		return nil, err
	}

	err := r.controller.CreateOrReplace(req.Context(), event)
	return event, err
}
