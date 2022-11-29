package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// RoleBindingsRouter handles requests for RoleBindings.
type RoleBindingsRouter struct {
	store storev2.Interface
}

// NewRoleBindingsRouter instantiates a new router for RoleBindings.
func NewRoleBindingsRouter(store storev2.Interface) *RoleBindingsRouter {
	return &RoleBindingsRouter{
		store: store,
	}
}

// Mount the RoleBindingsRouter on the given parent Router
func (r *RoleBindingsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:rolebindings}",
	}

	handlers := handlers.NewHandlers[*corev2.RoleBinding](r.store)

	routes.Del(handlers.DeleteResource)
	routes.Get(handlers.GetResource)
	routes.List(handlers.ListResources, corev2.RoleBindingFields)
	routes.ListAllNamespaces(handlers.ListResources, "/{resource:rolebindings}", corev2.RoleBindingFields)
	routes.Patch(handlers.PatchResource)
	routes.Post(handlers.CreateResource)
	routes.Put(handlers.CreateOrUpdateResource)
}
