package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// ClusterRoleBindingsRouter handles requests for ClusterRoleBindings.
type ClusterRoleBindingsRouter struct {
	handlers handlers.Handlers
}

// NewClusterRoleBindingsRouter instantiates a new router for ClusterRoleBindings.
func NewClusterRoleBindingsRouter(store storev2.Interface) *ClusterRoleBindingsRouter {
	return &ClusterRoleBindingsRouter{
		handlers: handlers.Handlers{
			Resource: &corev2.ClusterRoleBinding{},
			Store:    store,
		},
	}
}

// Mount the ClusterRoleBindingsRouter on the given parent Router
func (r *ClusterRoleBindingsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/{resource:clusterrolebindings}",
	}

	routes.Del(r.handlers.DeleteResource)
	routes.Get(r.handlers.GetResource)
	routes.List(r.handlers.ListResources, corev2.ClusterRoleBindingFields)
	routes.Patch(r.handlers.PatchResource)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
}
