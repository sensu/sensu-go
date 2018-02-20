package routers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/types"
)

// ChecksRouter handles requests for /checks
type ChecksRouter struct {
	controller actions.CheckController
}

// NewChecksRouter instantiates new router for controlling check resources
func NewChecksRouter(store queueStore) *ChecksRouter {
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

	// handlefunc returns a custom status and response
	parent.HandleFunc("/checks/{id}/execute", r.adhocRequest).Methods(http.MethodPost)
}

func (r *ChecksRouter) list(req *http.Request) (interface{}, error) {
	records, err := r.controller.Query(req.Context())
	return records, err
}

func (r *ChecksRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	record, err := r.controller.Find(req.Context(), id)
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
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	err = r.controller.Destroy(req.Context(), id)
	return nil, err
}

func (r *ChecksRouter) addCheckHook(req *http.Request) (interface{}, error) {
	cfg := types.HookList{}
	if err := unmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	err = r.controller.AddCheckHook(req.Context(), id, cfg)

	return nil, err
}

func (r *ChecksRouter) removeCheckHook(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	typ, err := url.PathUnescape(params["type"])
	if err != nil {
		return nil, err
	}
	hook, err := url.PathUnescape(params["hook"])
	if err != nil {
		return nil, err
	}
	err = r.controller.RemoveCheckHook(req.Context(), id, typ, hook)
	return nil, err
}

func (r *ChecksRouter) adhocRequest(w http.ResponseWriter, req *http.Request) {
	adhocReq := types.AdhocRequest{}
	if err := unmarshalBody(req, &adhocReq); err != nil {
		writeError(w, err)
		return
	}
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		writeError(w, err)
		return
	}
	if err := r.controller.QueueAdhocRequest(req.Context(), id, &adhocReq); err != nil {
		writeError(w, err)
		return
	}

	response := make(map[string]interface{})
	response["issued"] = time.Now().Unix()
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	if _, err := w.Write(jsonResponse); err != nil {
		writeError(w, err)
	}
}
