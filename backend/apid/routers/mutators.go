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

// MutatorsRouter handles /mutators requests.
type MutatorsRouter struct {
	controller actions.MutatorController
}

// NewMutatorsRouter creates a new MutatorsRouter.
func NewMutatorsRouter(store store.MutatorStore) *MutatorsRouter {
	return &MutatorsRouter{
		controller: actions.NewMutatorController(store),
	}
}

// Mount the MutatorsRouter to a parent Router
func (r *MutatorsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:mutators}",
	}
	routes.List(r.list)
	routes.Get(r.find)
	routes.Post(r.create)
	routes.Del(r.destroy)
	routes.Put(r.createOrReplace)
}

func (r *MutatorsRouter) list(req *http.Request) (interface{}, error) {
	return r.controller.Query(req.Context())
}

func (r *MutatorsRouter) listAllNamespaces(req *http.Request) (interface{}, error) {
	// Make sure the request context is empty so we query across all namespaces
	ctx := context.WithValue(req.Context(), types.NamespaceKey, "")

	return r.list(req.WithContext(ctx))
}

func (r *MutatorsRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	return r.controller.Find(req.Context(), id)
}

func (r *MutatorsRouter) create(req *http.Request) (interface{}, error) {
	mut := types.Mutator{}
	if err := UnmarshalBody(req, &mut); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), mut)
	return mut, err
}

func (r *MutatorsRouter) createOrReplace(req *http.Request) (interface{}, error) {
	mutator := types.Mutator{}
	if err := UnmarshalBody(req, &mutator); err != nil {
		return nil, err
	}

	return mutator, r.controller.CreateOrReplace(req.Context(), mutator)
}

func (r *MutatorsRouter) destroy(req *http.Request) (interface{}, error) {
	params := actions.QueryParams(mux.Vars(req))
	name, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	err = r.controller.Destroy(req.Context(), name)
	return nil, err
}
