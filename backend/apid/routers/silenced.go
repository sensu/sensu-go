package routers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// SilencedRouter handles requests for /users
type SilencedRouter struct {
	controller actions.SilencedController
}

// NewSilencedRouter instantiates new router for controlling user resources
func NewSilencedRouter(store store.Store) *SilencedRouter {
	return &SilencedRouter{
		controller: actions.NewSilencedController(store),
	}
}

// Mount the SilencedRouter to a parent Router
func (r *SilencedRouter) Mount(parent *mux.Router) {
	routes := resourceRoute{router: parent, pathPrefix: "/silenced"}
	routes.index(r.list)
	routes.show(r.find)
	routes.create(r.create)
	routes.update(r.update)
	routes.destroy(r.destroy)

	// Custom
	routes.path("subscriptions/{subscription}", r.list).Methods(http.MethodGet)
	routes.path("checks/{check}", r.list).Methods(http.MethodGet)
}

func (r *SilencedRouter) list(req *http.Request) (interface{}, error) {
	params := actions.QueryParams(mux.Vars(req))
	return r.controller.Query(req.Context(), params)
}

func (r *SilencedRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	return r.controller.Find(req.Context(), id)
}

func (r *SilencedRouter) create(req *http.Request) (interface{}, error) {
	cfg := types.Silenced{}
	if err := unmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), cfg)
	return cfg, err
}

func (r *SilencedRouter) update(req *http.Request) (interface{}, error) {
	cfg := types.Silenced{}
	if err := unmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	err := r.controller.Update(req.Context(), cfg)
	return cfg, err
}

func (r *SilencedRouter) destroy(req *http.Request) (interface{}, error) {
	params := actions.QueryParams(mux.Vars(req))
	err := r.controller.Destroy(req.Context(), params)
	return nil, err
}
