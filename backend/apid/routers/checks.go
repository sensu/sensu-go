package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/useractions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// ChecksRouter handles requests for /checks
type ChecksRouter struct {
	controller interface {
		useractions.Fetcher
		useractions.CheckMutator
		useractions.Destroyer
	}
}

// NewChecksRouter instantiates new checks controller
func NewChecksRouter(store store.CheckConfigStore) *ChecksRouter {
	return &ChecksRouter{
		controller: useractions.NewCheckController(store),
	}
}

// Mount the ChecksRouter to a parent Router
func (r *ChecksRouter) Mount(parent *mux.Router) {
	routes := resourceRoute{router: parent, pathPrefix: "/checks"}
	routes.index(r.list)
	routes.show(r.find)
	routes.create(r.create)
	routes.update(r.update)
	routes.destroy(r.destroy)
}

func (r *ChecksRouter) list(req *http.Request) (interface{}, error) {
	params := useractions.QueryParams(mux.Vars(req))
	records, err := r.controller.Query(req.Context(), params)
	return records, err
}

func (r *ChecksRouter) find(req *http.Request) (interface{}, error) {
	params := useractions.QueryParams(mux.Vars(req))
	record, err := r.controller.Find(req.Context(), params)
	return record, err
}

func (r *ChecksRouter) create(req *http.Request) (interface{}, error) {
	cfg := types.CheckConfig{}
	if err := unmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), cfg)
	return cfg, err
}

func (r *ChecksRouter) update(req *http.Request) (interface{}, error) {
	cfg := types.CheckConfig{}
	if err := unmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	err := r.controller.Update(req.Context(), cfg)
	return cfg, err
}

func (r *ChecksRouter) destroy(req *http.Request) (interface{}, error) {
	params := useractions.QueryParams(mux.Vars(req))
	err := r.controller.Destroy(req.Context(), params)
	return nil, err
}
