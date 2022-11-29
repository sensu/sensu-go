package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// ClusterRoleBindingsRouter handles requests for ClusterRoleBindings.
type ClusterRoleBindingsRouter struct {
	store storev2.Interface
}

// NewClusterRoleBindingsRouter instantiates a new router for ClusterRoleBindings.
func NewClusterRoleBindingsRouter(store storev2.Interface) *ClusterRoleBindingsRouter {
	return &ClusterRoleBindingsRouter{
		store: store,
	}
}

// Mount the ClusterRoleBindingsRouter on the given parent Router
func (r *ClusterRoleBindingsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/{resource:clusterrolebindings}",
	}

	handlers := handlers.NewHandlers[*corev2.ClusterRoleBinding](r.store)

	routes.Del(handlers.DeleteResource)
	routes.Get(handlers.GetResource)
	routes.List(handlers.ListResources, corev2.ClusterRoleBindingFields)
	routes.Patch(handlers.PatchResource)
	routes.Post(handlers.CreateResource)
	routes.Put(handlers.CreateOrUpdateResource)
}
