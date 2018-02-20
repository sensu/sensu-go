package routers

import (
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
	routes := resourceRoute{router: parent, pathPrefix: "/mutators"}
	routes.index(r.list)
	routes.show(r.find)
	routes.create(r.create)
	routes.update(r.update)
	routes.destroy(r.destroy)
}

func (r *MutatorsRouter) list(req *http.Request) (interface{}, error) {
	return r.controller.Query(req.Context())
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
	if err := unmarshalBody(req, &mut); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), mut)
	return mut, err
}

func (r *MutatorsRouter) update(req *http.Request) (interface{}, error) {
	mut := types.Mutator{}
	if err := unmarshalBody(req, &mut); err != nil {
		return nil, err
	}

	err := r.controller.Update(req.Context(), mut)
	return mut, err
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
