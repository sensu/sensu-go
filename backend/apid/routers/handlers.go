package routers

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// HandlersRouter handles requests for /handlers
type HandlersRouter struct {
	controller actions.HandlerController
}

// NewHandlersRouter instantiates new router for controlling handler resources
func NewHandlersRouter(store store.HandlerStore) *HandlersRouter {
	return &HandlersRouter{
		controller: actions.NewHandlerController(store),
	}
}

// Mount the HandlersRouter to a parent Router
func (r *HandlersRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:handlers}",
	}

	routes.Del(r.destroy)
	routes.Get(r.find)
	routes.List(r.list)
	routes.ListAllNamespaces(r.listAllNamespaces, "/{resource:handlers}")
	routes.Post(r.create)
	routes.Put(r.createOrReplace)
}

func (r *HandlersRouter) create(req *http.Request) (interface{}, error) {
	handler := types.Handler{}
	if err := UnmarshalBody(req, &handler); err != nil {
		return nil, err
	}

	return handler, r.controller.Create(req.Context(), handler)
}

func (r *HandlersRouter) createOrReplace(req *http.Request) (interface{}, error) {
	handler := types.Handler{}
	if err := UnmarshalBody(req, &handler); err != nil {
		return nil, err
	}

	return handler, r.controller.CreateOrReplace(req.Context(), handler)
}

func (r *HandlersRouter) destroy(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	return nil, r.controller.Destroy(req.Context(), id)
}

func (r *HandlersRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	return r.controller.Find(req.Context(), id)
}

func (r *HandlersRouter) list(req *http.Request) (interface{}, error) {
	return r.controller.Query(req.Context())
}

func (r *HandlersRouter) listAllNamespaces(req *http.Request) (interface{}, error) {
	// Make sure the request context is empty so we query across all namespaces
	ctx := context.WithValue(req.Context(), types.NamespaceKey, "")

	return r.list(req.WithContext(ctx))
}
