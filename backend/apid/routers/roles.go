package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// RolesRouter handles requests for Roles.
type RolesRouter struct {
	handlers handlers.Handlers
}

// NewRolesRouter instantiates a new router for Roles.
func NewRolesRouter(store storev2.Interface) *RolesRouter {
	return &RolesRouter{
		handlers: handlers.Handlers{
			Resource: &corev2.Role{},
			Store:    store,
		},
	}
}

// Mount the RolesRouter on the given parent Router
func (r *RolesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:roles}",
	}

	routes.Del(r.handlers.DeleteResource)
	routes.Get(r.handlers.GetResource)
	routes.List(r.handlers.ListResources, corev2.RoleFields)
	routes.ListAllNamespaces(r.handlers.ListResources, "/{resource:roles}", corev2.RoleFields)
	routes.Patch(r.handlers.PatchResource)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
}
