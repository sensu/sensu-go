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

// HooksRouter handles requests for /hooks
type HooksRouter struct {
	controller actions.HookController
}

// NewHooksRouter instantiates new router for controlling hook resources
func NewHooksRouter(store store.HookConfigStore) *HooksRouter {
	return &HooksRouter{
		controller: actions.NewHookController(store),
	}
}

// Mount the HooksRouter to a parent Router
func (r *HooksRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:hooks}",
	}

	routes.Del(r.destroy)
	routes.Get(r.find)
	routes.List(r.list)
	routes.ListAllNamespaces(r.list, "/{resource:hooks}")
	routes.Post(r.create)
	routes.Put(r.createOrReplace)
}

func (r *HooksRouter) list(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	records, continueToken, err := r.controller.Query(req.Context())

	if continueToken != "" {
		w.Header().Set(corev2.PaginationContinueHeader, continueToken)
	}

	return records, err
}

func (r *HooksRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	record, err := r.controller.Find(req.Context(), id)
	return record, err
}

func (r *HooksRouter) create(req *http.Request) (interface{}, error) {
	cfg := types.HookConfig{}
	if err := UnmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), cfg)
	return cfg, err
}

func (r *HooksRouter) createOrReplace(req *http.Request) (interface{}, error) {
	cfg := types.HookConfig{}
	if err := UnmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	err := r.controller.CreateOrReplace(req.Context(), cfg)
	return cfg, err
}

func (r *HooksRouter) destroy(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	err = r.controller.Destroy(req.Context(), id)
	return nil, err
}
