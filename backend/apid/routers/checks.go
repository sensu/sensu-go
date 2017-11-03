package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/useractions"
	"github.com/sensu/sensu-go/backend/store"
)

// ChecksRouter handles requests for /checks
type ChecksRouter struct {
	controller useractions.CheckActions
}

// NewChecksRouter instantiates new checks controller
func NewChecksRouter(store store.CheckConfigStore) *ChecksRouter {
	return &ChecksRouter{
		controller: useractions.NewCheckActions(nil, store),
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
	fetcher := r.controller.WithContext(req.Context())
	records, err := fetcher.Query(useractions.QueryParams{})
	return records, err
}

func (r *ChecksRouter) find(req *http.Request) (interface{}, error) {
	fetcher := r.controller.WithContext(req.Context())
	params := useractions.QueryParams(mux.Vars(req))
	record, err := fetcher.Find(params)
	return record, err
}

func (r *ChecksRouter) create(req *http.Request) (interface{}, error) {
	cfg := types.CheckConfig{}
	if err := unmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	mutator := r.controller.WithContext(req.Context())
	err := mutator.Create(cfg)

	return cfg, err
}

func (r *ChecksRouter) update(req *http.Request) (interface{}, error) {
	cfg := types.CheckConfig{}
	if err := unmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	mutator := r.controller.WithContext(req.Context())
	err := mutator.Update(cfg)

	return cfg, err
}

func (r *ChecksRouter) destroy(req *http.Request) (interface{}, error) {
	destoyer := r.controller.WithContext(req.Context())
	params := useractions.QueryParams(mux.Vars(req))
	return nil, destoyer.Destroy(params)
}
