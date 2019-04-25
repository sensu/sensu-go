package routers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// EventFiltersRouter handles /filters requests.
type EventFiltersRouter struct {
	controller actions.EventFilterController
}

// NewEventFiltersRouter creates a new EventFiltersRouter.
func NewEventFiltersRouter(store store.EventFilterStore) *EventFiltersRouter {
	return &EventFiltersRouter{
		controller: actions.NewEventFilterController(store),
	}
}

// Mount the EventFiltersRouter to a parent Router
func (r *EventFiltersRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:filters}",
	}

	routes.Del(r.destroy)
	routes.Get(r.find)
	routes.List(r.controller.List, corev2.EventFilterFields)
	routes.ListAllNamespaces(r.controller.List, "/{resource:filters}", corev2.EventFilterFields)
	routes.Post(r.create)
	routes.Put(r.createOrReplace)
}

func (r *EventFiltersRouter) find(req *http.Request) (interface{}, error) {
	params := actions.QueryParams(mux.Vars(req))
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	return r.controller.Find(req.Context(), id)
}

func (r *EventFiltersRouter) create(req *http.Request) (interface{}, error) {
	filter := types.EventFilter{}
	if err := UnmarshalBody(req, &filter); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), filter)
	return filter, err
}

func (r *EventFiltersRouter) createOrReplace(req *http.Request) (interface{}, error) {
	filter := types.EventFilter{}
	if err := UnmarshalBody(req, &filter); err != nil {
		return nil, err
	}

	err := r.controller.CreateOrReplace(req.Context(), filter)
	return filter, err
}

func (r *EventFiltersRouter) destroy(req *http.Request) (interface{}, error) {
	params := actions.QueryParams(mux.Vars(req))
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	err = r.controller.Destroy(req.Context(), id)
	return nil, err
}
