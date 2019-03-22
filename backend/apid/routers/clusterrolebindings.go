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

// ClusterRoleBindingsRouter handles requests for ClusterRoleBindings.
type ClusterRoleBindingsRouter struct {
	controller actions.ClusterRoleBindingController
}

// NewClusterRoleBindingsRouter instantiates a new router for ClusterRoleBindings.
func NewClusterRoleBindingsRouter(store store.ClusterRoleBindingStore) *ClusterRoleBindingsRouter {
	return &ClusterRoleBindingsRouter{
		controller: actions.NewClusterRoleBindingController(store),
	}
}

// Mount the ClusterRoleBindingsRouter on the given parent Router
func (r *ClusterRoleBindingsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/{resource:clusterrolebindings}",
	}
	routes.List(r.list)
	routes.Get(r.find)
	routes.Post(r.create)
	routes.Del(r.destroy)
	routes.Put(r.createOrReplace)
}

func (r *ClusterRoleBindingsRouter) create(req *http.Request) (interface{}, error) {
	obj := types.ClusterRoleBinding{}

	if err := UnmarshalBody(req, &obj); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), obj)
	return obj, err
}

func (r *ClusterRoleBindingsRouter) createOrReplace(req *http.Request) (interface{}, error) {
	obj := types.ClusterRoleBinding{}

	if err := UnmarshalBody(req, &obj); err != nil {
		return nil, err
	}

	err := r.controller.CreateOrReplace(req.Context(), obj)
	return obj, err
}

func (r *ClusterRoleBindingsRouter) destroy(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)

	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}

	err = r.controller.Destroy(req.Context(), id)
	return nil, err
}

func (r *ClusterRoleBindingsRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)

	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}

	obj, err := r.controller.Get(req.Context(), id)
	return obj, err
}

func (r *ClusterRoleBindingsRouter) list(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	objs, continueToken, err := r.controller.List(req.Context())

	if continueToken != "" {
		w.Header().Set(corev2.PaginationContinueHeader, continueToken)
	}

	return objs, err
}
