package routers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// ClusterRolesRouter handles requests for ClusterRoles.
type ClusterRolesRouter struct {
	controller actions.ClusterRoleController
}

// NewClusterRolesRouter instantiates a new router for ClusterRoles.
func NewClusterRolesRouter(store store.ClusterRoleStore) *ClusterRolesRouter {
	return &ClusterRolesRouter{
		controller: actions.NewClusterRoleController(store),
	}
}

// Mount the ClusterRolesRouter on the given parent Router
func (r *ClusterRolesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/{resource:clusterroles}",
	}
	routes.List(r.list)
	routes.Get(r.find)
	routes.Post(r.create)
	routes.Del(r.destroy)
	routes.Put(r.createOrReplace)
}

func (r *ClusterRolesRouter) create(req *http.Request) (interface{}, error) {
	obj := types.ClusterRole{}

	if err := UnmarshalBody(req, &obj); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), obj)
	return obj, err
}

func (r *ClusterRolesRouter) createOrReplace(req *http.Request) (interface{}, error) {
	obj := types.ClusterRole{}

	if err := UnmarshalBody(req, &obj); err != nil {
		return nil, err
	}

	err := r.controller.CreateOrReplace(req.Context(), obj)
	return obj, err
}

func (r *ClusterRolesRouter) destroy(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)

	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}

	err = r.controller.Destroy(req.Context(), id)
	return nil, err
}

func (r *ClusterRolesRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)

	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}

	obj, err := r.controller.Get(req.Context(), id)
	return obj, err
}

func (r *ClusterRolesRouter) list(req *http.Request) (interface{}, error) {
	objs, err := r.controller.List(req.Context())
	return objs, err
}
