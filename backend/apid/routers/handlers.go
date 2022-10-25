package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// HandlersRouter handles requests for /handlers
type HandlersRouter struct {
	handlers handlers.Handlers
}

// NewHandlersRouter instantiates new router for controlling handler resources
func NewHandlersRouter(store storev2.Interface) *HandlersRouter {
	return &HandlersRouter{
		handlers: handlers.Handlers{
			Resource: &corev2.Handler{},
			Store:    store,
		},
	}
}

// Mount the HandlersRouter to a parent Router
func (r *HandlersRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:handlers}",
	}
	routes.Del(r.handlers.DeleteResource)
	routes.Get(r.handlers.GetResource)
	routes.List(r.handlers.ListResources, corev2.HandlerFields)
	routes.ListAllNamespaces(r.handlers.ListResources, "/{resource:handlers}", corev2.HandlerFields)
	routes.Patch(r.handlers.PatchResource)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
}
