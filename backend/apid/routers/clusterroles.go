package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// ClusterRolesRouter handles requests for ClusterRoles.
type ClusterRolesRouter struct {
	store storev2.Interface
}

// NewClusterRolesRouter instantiates a new router for ClusterRoles.
func NewClusterRolesRouter(store storev2.Interface) *ClusterRolesRouter {
	return &ClusterRolesRouter{
		store: store,
	}
}

// Mount the ClusterRolesRouter on the given parent Router
func (r *ClusterRolesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/{resource:clusterroles}",
	}

	handlers := handlers.NewHandlers[*corev2.ClusterRole](r.store)

	routes.Del(handlers.DeleteResource)
	routes.Get(handlers.GetResource)
	routes.List(handlers.ListResources, corev3.ClusterRoleFields)
	routes.Patch(handlers.PatchResource)
	routes.Post(handlers.CreateResource)
	routes.Put(handlers.CreateOrUpdateResource)
}
