package routers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// checkController represents the controller needs of the ChecksRouter.
type checkController interface {
	AddCheckHook(context.Context, string, corev2.HookList) error
	RemoveCheckHook(context.Context, string, string, string) error
	QueueAdhocRequest(context.Context, string, *corev2.AdhocRequest) error
}

// ChecksRouter handles requests for /checks
type ChecksRouter struct {
	controller checkController
	handlers   handlers.Handlers
}

// NewChecksRouter instantiates new router for controlling check resources
func NewChecksRouter(store store.Store, getter types.QueueGetter) *ChecksRouter {
	return &ChecksRouter{
		controller: actions.NewCheckController(store, getter),
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
	routes.List(r.handlers.ListResources, corev2.CheckConfigFields)
	routes.ListAllNamespaces(r.handlers.ListResources, "/{resource:checks}", corev2.CheckConfigFields)
	routes.Patch(r.handlers.PatchResource)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)

	// Custom
	routes.Path("{id}/hooks/{type}", r.addCheckHook).Methods(http.MethodPut)
	routes.Path("{id}/hooks/{type}/hook/{hook}", r.removeCheckHook).Methods(http.MethodDelete)

	// handlefunc returns a custom status and response
	parent.HandleFunc(path.Join(routes.PathPrefix, "{id}/execute"), r.adhocRequest).Methods(http.MethodPost)
}

func (r *ChecksRouter) addCheckHook(req *http.Request) (interface{}, error) {
	cfg := corev2.HookList{}
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
	adhocReq := corev2.AdhocRequest{}
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
