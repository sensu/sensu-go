package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/store"
)

// NamespacesRouter handles requests for /namespaces
type NamespacesRouter struct {
	handlers handlers.Handlers
}

// NewNamespacesRouter instantiates new router for controlling check resources
func NewNamespacesRouter(store store.ResourceStore) *NamespacesRouter {
	return &NamespacesRouter{
		handlers: handlers.Handlers{
			Resource: &corev2.Namespace{},
			Store:    store,
		},
	}
}

// Mount the NamespacesRouter to a parent Router
func (r *NamespacesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/{resource:namespaces}",
	}

	routes.Del(r.handlers.DeleteResource)
	routes.Get(r.handlers.GetResource)
	routes.List(r.handlers.ListResources, corev2.NamespaceFields)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
}
