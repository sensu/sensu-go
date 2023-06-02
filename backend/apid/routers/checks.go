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
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/apid/request"
	"github.com/sensu/sensu-go/backend/queue"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
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
	store      storev2.Interface
}

// NewChecksRouter instantiates new router for controlling check resources
func NewChecksRouter(store storev2.Interface, queue queue.Client) *ChecksRouter {
	return &ChecksRouter{
		controller: actions.NewCheckController(store, queue),
		store:      store,
	}
}

// Mount the ChecksRouter to a parent Router
func (r *ChecksRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:checks}",
	}

	handlers := handlers.NewHandlers[*corev2.CheckConfig](r.store)

	routes.Del(handlers.DeleteResource)
	routes.Get(handlers.GetResource)
	routes.List(handlers.ListResources, corev3.CheckConfigFields)
	routes.ListAllNamespaces(handlers.ListResources, "/{resource:checks}", corev3.CheckConfigFields)
	routes.Patch(handlers.PatchResource)
	routes.Post(handlers.CreateResource)
	routes.Put(handlers.CreateOrUpdateResource)

	// Custom
	routes.Path("{id}/hooks/{type}", r.addCheckHook).Methods(http.MethodPut)
	routes.Path("{id}/hooks/{type}/hook/{hook}", r.removeCheckHook).Methods(http.MethodDelete)

	// handlefunc returns a custom status and response
	parent.HandleFunc(path.Join(routes.PathPrefix, "{id}/execute"), r.adhocRequest).Methods(http.MethodPost)
}

func (r *ChecksRouter) addCheckHook(req *http.Request) (handlers.HandlerResponse, error) {
	var cfg corev2.HookList
	var response handlers.HandlerResponse

	if err := json.NewDecoder(req.Body).Decode(&cfg); err != nil {
		return response, actions.NewError(actions.InvalidArgument, err)
	}

	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return response, err
	}
	err = r.controller.AddCheckHook(req.Context(), id, cfg)

	return response, err
}

func (r *ChecksRouter) removeCheckHook(req *http.Request) (handlers.HandlerResponse, error) {
	var response handlers.HandlerResponse

	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return response, err
	}
	typ, err := url.PathUnescape(params["type"])
	if err != nil {
		return response, err
	}
	hook, err := url.PathUnescape(params["hook"])
	if err != nil {
		return response, err
	}
	err = r.controller.RemoveCheckHook(req.Context(), id, typ, hook)
	return response, err
}

func (r *ChecksRouter) adhocRequest(w http.ResponseWriter, req *http.Request) {
	adhocReq, err := request.Resource[*corev2.AdhocRequest](req)
	if err != nil {
		WriteError(w, err)
		return
	}
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		WriteError(w, err)
		return
	}
	if err := r.controller.QueueAdhocRequest(req.Context(), id, adhocReq); err != nil {
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
