package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// EventFiltersRouter handles /filters requests.
type EventFiltersRouter struct {
	store storev2.Interface
}

// NewEventFiltersRouter creates a new EventFiltersRouter.
func NewEventFiltersRouter(store storev2.Interface) *EventFiltersRouter {
	return &EventFiltersRouter{
		store: store,
	}
}

// Mount the EventFiltersRouter to a parent Router
func (r *EventFiltersRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:filters}",
	}

	handlers := handlers.NewHandlers[*corev2.EventFilter](r.store)

	routes.Del(handlers.DeleteResource)
	routes.Get(handlers.GetResource)
	routes.List(handlers.ListResources, corev3.EventFilterFields)
	routes.ListAllNamespaces(handlers.ListResources, "/{resource:filters}", corev3.EventFilterFields)
	routes.Patch(handlers.PatchResource)
	routes.Post(handlers.CreateResource)
	routes.Put(handlers.CreateOrUpdateResource)
}
