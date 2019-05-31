package routers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// CheckController represents the controller needs of the ChecksRouter.
type CheckController interface {
	Create(context.Context, types.CheckConfig) error
	CreateOrReplace(context.Context, types.CheckConfig) error
	List(context.Context, *store.SelectionPredicate) ([]corev2.Resource, error)
	Find(context.Context, string) (*types.CheckConfig, error)
	Destroy(context.Context, string) error
	AddCheckHook(context.Context, string, types.HookList) error
	RemoveCheckHook(context.Context, string, string, string) error
	QueueAdhocRequest(context.Context, string, *types.AdhocRequest) error
}

// ChecksRouter handles requests for /checks
type ChecksRouter struct {
	controller CheckController
}

// NewChecksRouter instantiates new router for controlling check resources
func NewChecksRouter(ctrl CheckController) *ChecksRouter {
	return &ChecksRouter{
		controller: ctrl,
	}
}

// Mount the ChecksRouter to a parent Router
func (r *ChecksRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:checks}",
	}

	routes.Del(r.destroy)
	routes.Get(r.find)
	routes.List(r.controller.List, corev2.CheckConfigFields)
	routes.ListAllNamespaces(r.controller.List, "/{resource:checks}", corev2.CheckConfigFields)
	routes.Post(r.create)
	routes.Put(r.createOrReplace)

	// Custom
	routes.Path("{id}/hooks/{type}", r.addCheckHook).Methods(http.MethodPut)
	routes.Path("{id}/hooks/{type}/hook/{hook}", r.removeCheckHook).Methods(http.MethodDelete)

	// handlefunc returns a custom status and response
	parent.HandleFunc(path.Join(routes.PathPrefix, "{id}/execute"), r.adhocRequest).Methods(http.MethodPost)
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
	if err := UnmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), cfg)
	return cfg, err
}

func (r *ChecksRouter) createOrReplace(req *http.Request) (interface{}, error) {
	cfg := types.CheckConfig{}
	if err := UnmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	if err := checkMeta(&cfg, mux.Vars(req)); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	err := r.controller.CreateOrReplace(req.Context(), cfg)
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
	if err := UnmarshalBody(req, &cfg); err != nil {
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
	if err := UnmarshalBody(req, &adhocReq); err != nil {
		WriteError(w, err)
		return
	}
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		WriteError(w, err)
		return
	}
	if err := r.controller.QueueAdhocRequest(req.Context(), id, &adhocReq); err != nil {
		WriteError(w, err)
		return
	}

	response := make(map[string]interface{})
	response["issued"] = time.Now().Unix()
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	if _, err := w.Write(jsonResponse); err != nil {
		WriteError(w, err)
	}
}
