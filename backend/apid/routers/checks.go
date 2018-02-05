package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// ChecksRouter handles requests for /checks
type ChecksRouter struct {
	controller actions.CheckController
}

// NewChecksRouter instantiates new router for controlling check resources
func NewChecksRouter(store store.Store) *ChecksRouter {
	return &ChecksRouter{
		controller: actions.NewCheckController(store),
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

	// Custom
	routes.path("{id}/hooks/{type}", r.addCheckHook).Methods(http.MethodPut)
	routes.path("{id}/hooks/{type}/hook/{hook}", r.removeCheckHook).Methods(http.MethodDelete)
	routes.path("{id}/execute", r.adhocRequest).Methods(http.MethodPost)
}

func (r *ChecksRouter) list(req *http.Request) (interface{}, error) {
	records, err := r.controller.Query(req.Context())
	return records, err
}

func (r *ChecksRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	record, err := r.controller.Find(req.Context(), params["id"])
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
	params := mux.Vars(req)
	err := r.controller.Destroy(req.Context(), params["id"])
	return nil, err
}

func (r *ChecksRouter) addCheckHook(req *http.Request) (interface{}, error) {
	cfg := types.HookList{}
	if err := unmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	params := mux.Vars(req)
	err := r.controller.AddCheckHook(req.Context(), params["id"], cfg)

	return nil, err
}

func (r *ChecksRouter) removeCheckHook(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	err := r.controller.RemoveCheckHook(req.Context(), params["id"], params["type"], params["hook"])
	return nil, err
}

func (r *ChecksRouter) adhocRequest(req *http.Request) (interface{}, error) {
	adhocReq := types.AdhocRequest{}
	if err := unmarshalBody(req, &adhocReq); err != nil {
		return nil, err
	}
	params := mux.Vars(req)
	err := r.controller.QueueAdhocRequest(req.Context(), params["id"], adhocReq)
	// needs to return a 202 and json with {"issued":<timestamp>} to be
	// backwards compatible
	return nil, err
}
