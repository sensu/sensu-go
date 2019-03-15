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
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:silenced}",
	}

	routes.Del(r.destroy)
	routes.Get(r.find)
	routes.List(r.list)
	routes.ListAllNamespaces(r.list, "/{resource:silenced}")
	routes.Post(r.create)
	routes.Put(r.createOrReplace)

	// Custom
	routes.Router.HandleFunc("subscriptions/{subscriptions}", listHandler(r.list)).Methods(http.MethodGet)
	routes.Router.HandleFunc("checks/{check}", listHandler(r.list)).Methods(http.MethodGet)
}

func (r *SilencedRouter) list(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	return r.controller.Query(req.Context(), params["subscription"], params["check"])
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
	if err := UnmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), &cfg)
	return cfg, err
}

func (r *SilencedRouter) createOrReplace(req *http.Request) (interface{}, error) {
	cfg := types.Silenced{}
	if err := UnmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	err := r.controller.CreateOrReplace(req.Context(), cfg)
	return cfg, err
}

func (r *SilencedRouter) destroy(req *http.Request) (interface{}, error) {
	params := actions.QueryParams(mux.Vars(req))
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}

	err = r.controller.Destroy(req.Context(), id)
	return nil, err
}
