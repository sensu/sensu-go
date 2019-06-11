package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/store"
)

// ExtensionsRouter handles requests for /extensions
type ExtensionsRouter struct {
	handlers handlers.Handlers
}

// NewExtensionsRouter creates a new router for controlling extension resources
func NewExtensionsRouter(store store.ResourceStore) *ExtensionsRouter {
	return &ExtensionsRouter{
		handlers: handlers.Handlers{
			Resource: &corev2.Extension{},
			Store:    store,
		},
	}
}

// Mount the ExtensionsRouter to a parent Router
func (r *ExtensionsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:extensions}",
	}

	routes.Del(r.handlers.DeleteResource)
	routes.Get(r.handlers.GetResource)
	routes.List(r.handlers.ListResources, corev2.ExtensionFields)
	routes.ListAllNamespaces(r.handlers.ListResources, "/{resource:extensions}", corev2.ExtensionFields)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
}
