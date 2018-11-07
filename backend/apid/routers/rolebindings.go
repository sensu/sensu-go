package routers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// RoleBindingsRouter handles requests for RoleBindings.
type RoleBindingsRouter struct {
	controller actions.RoleBindingController
}

// NewRoleBindingsRouter instantiates a new router for RoleBindings.
func NewRoleBindingsRouter(store store.RoleBindingStore) *RoleBindingsRouter {
	return &RoleBindingsRouter{
		controller: actions.NewRoleBindingController(store),
	}
}

// Mount the RoleBindingsRouter on the given parent Router
func (r *RoleBindingsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{Router: parent, PathPrefix: "/apis/rbac/v2/namespaces/{namespace}/rolebindings"}
	routes.GetAll(r.list)
	routes.Get(r.find)
	routes.Post(r.create)
	routes.Del(r.destroy)
	routes.Put(r.createOrReplace)
}

func (r *RoleBindingsRouter) create(req *http.Request) (interface{}, error) {
	obj := types.RoleBinding{}

	if err := UnmarshalBody(req, &obj); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), obj)
	return obj, err
}

func (r *RoleBindingsRouter) createOrReplace(req *http.Request) (interface{}, error) {
	obj := types.RoleBinding{}

	if err := UnmarshalBody(req, &obj); err != nil {
		return nil, err
	}

	err := r.controller.CreateOrReplace(req.Context(), obj)
	return obj, err
}

func (r *RoleBindingsRouter) destroy(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)

	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}

	err = r.controller.Destroy(req.Context(), id)
	return nil, err
}

func (r *RoleBindingsRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)

	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}

	obj, err := r.controller.Get(req.Context(), id)
	return obj, err
}

func (r *RoleBindingsRouter) list(req *http.Request) (interface{}, error) {
	objs, err := r.controller.List(req.Context())
	return objs, err
}
