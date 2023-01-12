package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// HooksRouter handles requests for /hooks
type HooksRouter struct {
	store storev2.Interface
}

// NewHooksRouter instantiates new router for controlling hook resources
func NewHooksRouter(store storev2.Interface) *HooksRouter {
	return &HooksRouter{
		store: store,
	}
}

// Mount the HooksRouter to a parent Router
func (r *HooksRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:hooks}",
	}

	handlers := handlers.NewHandlers[*corev2.HookConfig](r.store)

	routes.Del(handlers.DeleteResource)
	routes.Get(handlers.GetResource)
	routes.List(handlers.ListResources, corev3.HookConfigFields)
	routes.ListAllNamespaces(handlers.ListResources, "/{resource:hooks}", corev3.HookConfigFields)
	routes.Patch(handlers.PatchResource)
	routes.Post(handlers.CreateResource)
	routes.Put(handlers.CreateOrUpdateResource)
}
