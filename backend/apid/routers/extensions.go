package routers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// ExtensionsRouter handles requests for /extensions
type ExtensionsRouter struct {
	controller actions.ExtensionController
}

// NewExtensionRouter instantiates new router for controlling extension resources
func NewExtensionRouter(store store.ExtensionRegistry) *ExtensionsRouter {
	return &ExtensionsRouter{
		controller: actions.NewExtensionController(store),
	}
}

// Mount the ExtensionsRouter to a parent Router
func (r *ExtensionsRouter) Mount(parent *mux.Router) {
	routes := resourceRoute{router: parent, pathPrefix: "/extensions"}
	routes.getAll(r.list)
	routes.get(r.find)
	routes.put(r.register)
	routes.del(r.deregister)
}

func (r *ExtensionsRouter) list(req *http.Request) (interface{}, error) {
	records, err := r.controller.Query(req.Context())
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
	if err := unmarshalBody(req, &extension); err != nil {
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
