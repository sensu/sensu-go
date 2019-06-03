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
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// CheckController represents the controller needs of the ChecksRouter.
type CheckController interface {
	List(context.Context, *store.SelectionPredicate) ([]corev2.Resource, error)
	AddCheckHook(context.Context, string, types.HookList) error
	RemoveCheckHook(context.Context, string, string, string) error
	QueueAdhocRequest(context.Context, string, *types.AdhocRequest) error
}

// ChecksRouter handles requests for /checks
type ChecksRouter struct {
	checkController CheckController
	handlers        handlers.Handlers
}

// NewChecksRouter instantiates new router for controlling check resources
func NewChecksRouter(checkController CheckController, store store.ResourceStore) *ChecksRouter {
	return &ChecksRouter{
		checkController: checkController,
		handlers: handlers.Handlers{
			Resource: &corev2.CheckConfig{},
			Store:    store,
		},
	}
}

// Mount the ChecksRouter to a parent Router
func (r *ChecksRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:checks}",
	}

	routes.Del(r.handlers.DeleteResource)
	routes.Get(r.handlers.GetResource)
	routes.List(r.handlers.List, corev2.CheckConfigFields)
	routes.ListAllNamespaces(r.handlers.List, "/{resource:checks}", corev2.CheckConfigFields)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)

	// Custom
	routes.Path("{id}/hooks/{type}", r.addCheckHook).Methods(http.MethodPut)
	routes.Path("{id}/hooks/{type}/hook/{hook}", r.removeCheckHook).Methods(http.MethodDelete)

	// handlefunc returns a custom status and response
	parent.HandleFunc(path.Join(routes.PathPrefix, "{id}/execute"), r.adhocRequest).Methods(http.MethodPost)
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
	err = r.checkController.AddCheckHook(req.Context(), id, cfg)

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
	err = r.checkController.RemoveCheckHook(req.Context(), id, typ, hook)
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
	if err := r.checkController.QueueAdhocRequest(req.Context(), id, &adhocReq); err != nil {
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
