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

	routes.Del(r.destroy)
	routes.Get(r.find)
	routes.List(r.list)
	routes.ListAllNamespaces(r.list, "/{resource:mutators}")
	routes.Post(r.create)
	routes.Put(r.createOrReplace)
}

func (r *MutatorsRouter) list(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	records, continueToken, err := r.controller.Query(req.Context())

	if continueToken != "" {
		w.Header().Set(corev2.PaginationContinueHeader, continueToken)
	}

	return records, err
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
