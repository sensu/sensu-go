package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/store"
)

// HooksRouter handles requests for /hooks
type HooksRouter struct {
	handlers handlers.Handlers
}

// NewHooksRouter instantiates new router for controlling hook resources
func NewHooksRouter(store store.ResourceStore) *HooksRouter {
	return &HooksRouter{
		handlers: handlers.Handlers{
			Resource: &corev2.HookConfig{},
			Store:    store,
		},
	}
}

// Mount the HooksRouter to a parent Router
func (r *HooksRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:hooks}",
	}

	routes.Del(r.handlers.DeleteResource)
	routes.Get(r.handlers.GetResource)
	routes.List(r.handlers.ListResources, corev2.HookConfigFields)
	routes.ListAllNamespaces(r.handlers.ListResources, "/{resource:hooks}", corev2.HookConfigFields)
	routes.Patch(r.handlers.PatchResource)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
}
