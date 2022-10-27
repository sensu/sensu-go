package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// RoleBindingsRouter handles requests for RoleBindings.
type RoleBindingsRouter struct {
	handlers handlers.Handlers
}

// NewRoleBindingsRouter instantiates a new router for RoleBindings.
func NewRoleBindingsRouter(store storev2.Interface) *RoleBindingsRouter {
	return &RoleBindingsRouter{
		handlers: handlers.Handlers{
			Resource: &corev2.RoleBinding{},
			Store:    store,
		},
	}
}

// Mount the RoleBindingsRouter on the given parent Router
func (r *RoleBindingsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:rolebindings}",
	}

	routes.Del(r.handlers.DeleteResource)
	routes.Get(r.handlers.GetResource)
	routes.List(r.handlers.ListResources, corev2.RoleBindingFields)
	routes.ListAllNamespaces(r.handlers.ListResources, "/{resource:rolebindings}", corev2.RoleBindingFields)
	routes.Patch(r.handlers.PatchResource)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
}
