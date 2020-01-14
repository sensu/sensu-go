package routers

import (
	"context"
	"net/http"
	"net/url"
	"path"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
)

// EventsRouter handles requests for /events
type EventsRouter struct {
	controller eventController
}

// eventController represents the controller needs of the EventsRouter.
type eventController interface {
	CreateOrReplace(ctx context.Context, check *corev2.Event) error
	Delete(ctx context.Context, entity, check string) error
	Get(ctx context.Context, entity, check string) (*corev2.Event, error)
	List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error)
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
	routes.Path("{entity}/{check}", r.get).Methods(http.MethodGet)
	routes.Path("{entity}/{check}", r.delete).Methods(http.MethodDelete)
	routes.Path("{entity}/{check}", r.createOrReplace).Methods(http.MethodPost, http.MethodPut)

	// Additionaly allow a subcollection to be specified when listing events,
	// which correspond to the entity name here
	parent.HandleFunc(path.Join(routes.PathPrefix, "{subcollection}"),
		listerHandler(r.controller.List, corev2.EventFields)).Methods(http.MethodGet)
}

func (r *EventsRouter) get(req *http.Request) (interface{}, error) {
	params := actions.QueryParams(mux.Vars(req))
	entity := url.PathEscape(params["entity"])
	check := url.PathEscape(params["check"])
	record, err := r.controller.Get(req.Context(), entity, check)
	return record, err
}

func (r *EventsRouter) delete(req *http.Request) (interface{}, error) {
	params := actions.QueryParams(mux.Vars(req))
	entity := url.PathEscape(params["entity"])
	check := url.PathEscape(params["check"])
	return nil, r.controller.Delete(req.Context(), entity, check)
}

func (r *EventsRouter) create(req *http.Request) (interface{}, error) {
	event := &corev2.Event{}
	if err := UnmarshalBody(req, event); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	vars := mux.Vars(req)

	if event.Entity != nil {
		if err := handlers.MetaPathValues(event.Entity, vars, "entity"); err != nil {
			return nil, err
		}

		if err := handlers.CheckMeta(event.Entity, vars); err != nil {
			return nil, actions.NewError(actions.InvalidArgument, err)
		}
	}

	if event.Check != nil {
		if err := handlers.MetaPathValues(event.Check, vars, "check"); err != nil {
			return nil, err
		}

		if err := handlers.CheckMeta(event.Check, vars); err != nil {
			return nil, actions.NewError(actions.InvalidArgument, err)
		}
	}

	err := r.controller.CreateOrReplace(req.Context(), event)
	return nil, err
}

func (r *EventsRouter) createOrReplace(req *http.Request) (interface{}, error) {
	event := &corev2.Event{}
	if err := UnmarshalBody(req, event); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	vars := mux.Vars(req)

	if event.Entity != nil {
		if err := handlers.MetaPathValues(event.Entity, vars, "entity"); err != nil {
			return nil, err
		}

		if err := handlers.CheckMeta(event.Entity, vars); err != nil {
			return nil, actions.NewError(actions.InvalidArgument, err)
		}
	}

	if event.Check != nil {
		if err := handlers.MetaPathValues(event.Check, vars, "check"); err != nil {
			return nil, err
		}

		if err := handlers.CheckMeta(event.Check, vars); err != nil {
			return nil, actions.NewError(actions.InvalidArgument, err)
		}
	}

	err := r.controller.CreateOrReplace(req.Context(), event)
	return nil, err
}
