package routers

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
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
	routes.List(r.list)
	routes.Get(r.find)
	routes.Post(r.create)
	routes.Del(r.destroy)
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

func (r *RolesRouter) list(req *http.Request) (interface{}, error) {
	objs, err := r.controller.List(req.Context())
	return objs, err
}

func (r *RolesRouter) listAllNamespaces(req *http.Request) (interface{}, error) {
	// Make sure the request context is empty so we query across all namespaces
	ctx := context.WithValue(req.Context(), types.NamespaceKey, "")

	return r.list(req.WithContext(ctx))
}
