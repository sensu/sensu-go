package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/internal/apis/rbac"
	storev2 "github.com/sensu/sensu-go/storage"
)

// RBACRouter handles requests for RBAC resources
type RBACRouter struct {
	controller actions.GenericController
}

// NewRBACRouter instantiates new router for controlling RBAC resources
func NewRBACRouter(store storev2.Store) *RBACRouter {
	return &RBACRouter{
		controller: actions.NewGenericController(store),
	}
}

// Mount the RolesRouter to a parent Router
func (r *RBACRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{Router: parent, PathPrefix: "/apis/rbac/v1alpha1"}
	routes.Path("clusterroles", r.list).Methods(http.MethodGet, http.MethodDelete)
}

func (r *RBACRouter) list(req *http.Request) (interface{}, error) {
	clusterRoles := []rbac.ClusterRole{}
	err := r.controller.List(req.Context(), "clusterroles", &clusterRoles)
	return clusterRoles, err
}
