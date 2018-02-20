package routers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
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
	routes := resourceRoute{router: parent, pathPrefix: "/filters"}
	routes.index(r.list)
	routes.show(r.find)
	routes.create(r.create)
	routes.update(r.update)
	routes.destroy(r.destroy)
}

func (r *EventFiltersRouter) list(req *http.Request) (interface{}, error) {
	return r.controller.Query(req.Context())
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
	if err := unmarshalBody(req, &filter); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), filter)
	return filter, err
}

func (r *EventFiltersRouter) update(req *http.Request) (interface{}, error) {
	filter := types.EventFilter{}
	if err := unmarshalBody(req, &filter); err != nil {
		return nil, err
	}

	err := r.controller.Update(req.Context(), filter)
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
