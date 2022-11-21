package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// HandlersRouter handles requests for /handlers
type HandlersRouter struct {
	store storev2.Interface
}

// NewHandlersRouter instantiates new router for controlling handler resources
func NewHandlersRouter(store storev2.Interface) *HandlersRouter {
	return &HandlersRouter{
		store: store,
	}
}

// Mount the HandlersRouter to a parent Router
func (r *HandlersRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:handlers}",
	}
	handlers := handlers.NewHandlers[*corev2.Handler](r.store)
	routes.Del(handlers.DeleteResource)
	routes.Get(handlers.GetResource)
	routes.List(handlers.ListResources, corev2.HandlerFields)
	routes.ListAllNamespaces(handlers.ListResources, "/{resource:handlers}", corev2.HandlerFields)
	routes.Patch(handlers.PatchResource)
	routes.Post(handlers.CreateResource)
	routes.Put(handlers.CreateOrUpdateResource)
}
