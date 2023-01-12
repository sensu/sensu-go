package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// RolesRouter handles requests for Roles.
type RolesRouter struct {
	store storev2.Interface
}

// NewRolesRouter instantiates a new router for Roles.
func NewRolesRouter(store storev2.Interface) *RolesRouter {
	return &RolesRouter{
		store: store,
	}
}

// Mount the RolesRouter on the given parent Router
func (r *RolesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:roles}",
	}

	handlers := handlers.NewHandlers[*corev2.Role](r.store)

	routes.Del(handlers.DeleteResource)
	routes.Get(handlers.GetResource)
	routes.List(handlers.ListResources, corev3.RoleFields)
	routes.ListAllNamespaces(handlers.ListResources, "/{resource:roles}", corev3.RoleFields)
	routes.Patch(handlers.PatchResource)
	routes.Post(handlers.CreateResource)
	routes.Put(handlers.CreateOrUpdateResource)
}
