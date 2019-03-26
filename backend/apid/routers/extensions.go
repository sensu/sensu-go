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

// ExtensionsRouter handles requests for /extensions
type ExtensionsRouter struct {
	controller actions.ExtensionController
}

// NewExtensionsRouter creates a new router for controlling extension resources
func NewExtensionsRouter(store store.ExtensionRegistry) *ExtensionsRouter {
	return &ExtensionsRouter{
		controller: actions.NewExtensionController(store),
	}
}

// Mount the ExtensionsRouter to a parent Router
func (r *ExtensionsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:extensions}",
	}

	routes.Del(r.deregister)
	routes.Get(r.find)
	routes.List(r.list)
	routes.ListAllNamespaces(r.list, "/{resource:extensions}")
	routes.Put(r.register)
}

func (r *ExtensionsRouter) list(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	records, continueToken, err := r.controller.Query(req.Context())

	if continueToken != "" {
		w.Header().Set(corev2.PaginationContinueHeader, continueToken)
	}

	return records, err
}

func (r *ExtensionsRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	extensionPath, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	record, err := r.controller.Find(req.Context(), extensionPath)
	return record, err
}

func (r *ExtensionsRouter) register(req *http.Request) (interface{}, error) {
	var extension types.Extension
	if err := UnmarshalBody(req, &extension); err != nil {
		return nil, err
	}
	err := r.controller.Register(req.Context(), extension)
	return extension, err
}

func (r *ExtensionsRouter) deregister(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	extensionPath, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	return nil, r.controller.Deregister(req.Context(), extensionPath)
}
