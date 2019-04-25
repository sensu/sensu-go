package routers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// RolesRouter handles requests for Roles.
type RolesRouter struct {
	controller actions.RoleController
}

// NewRolesRouter instantiates a new router for Roles.
func NewRolesRouter(store store.RoleStore) *RolesRouter {
	return &RolesRouter{
		controller: actions.NewRoleController(store),
	}
}

// Mount the RolesRouter on the given parent Router
func (r *RolesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:roles}",
	}

	routes.Del(r.destroy)
	routes.Get(r.find)
	routes.List(r.controller.List, corev2.RoleFields)
	routes.ListAllNamespaces(r.controller.List, "/{resource:roles}", corev2.RoleFields)
	routes.Post(r.create)
	routes.Put(r.createOrReplace)
}

func (r *RolesRouter) create(req *http.Request) (interface{}, error) {
	obj := types.Role{}

	if err := UnmarshalBody(req, &obj); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), obj)
	return obj, err
}

func (r *RolesRouter) createOrReplace(req *http.Request) (interface{}, error) {
	obj := types.Role{}

	if err := UnmarshalBody(req, &obj); err != nil {
		return nil, err
	}

	err := r.controller.CreateOrReplace(req.Context(), obj)
	return obj, err
}

func (r *RolesRouter) destroy(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)

	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}

	err = r.controller.Destroy(req.Context(), id)
	return nil, err
}

func (r *RolesRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)

	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}

	obj, err := r.controller.Get(req.Context(), id)
	return obj, err
}
