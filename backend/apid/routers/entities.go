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

// EntitiesRouter handles requests for /entities
type EntitiesRouter struct {
	controller actions.EntityController
}

// NewEntitiesRouter instantiates new router for controlling entities resources
func NewEntitiesRouter(store store.EntityStore) *EntitiesRouter {
	return &EntitiesRouter{
		controller: actions.NewEntityController(store),
	}
}

// Mount the EntitiesRouter to a parent Router
func (r *EntitiesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:entities}",
	}

	routes.Del(r.destroy)
	routes.Get(r.find)
	routes.List(r.list)
	routes.ListAllNamespaces(r.listAllNamespaces, "/{resource:entities}")
	routes.Post(r.create)
	routes.Put(r.createOrReplace)
}

func (r *EntitiesRouter) destroy(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	err = r.controller.Destroy(req.Context(), id)
	return nil, err
}

func (r *EntitiesRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	record, err := r.controller.Find(req.Context(), id)
	return record, err
}

func (r *EntitiesRouter) list(req *http.Request) (interface{}, error) {
	records, err := r.controller.Query(req.Context())
	return records, err
}

func (r *EntitiesRouter) listAllNamespaces(req *http.Request) (interface{}, error) {
	// Make sure the request context is empty so we query across all namespaces
	ctx := context.WithValue(req.Context(), types.NamespaceKey, "")

	return r.list(req.WithContext(ctx))
}

func (r *EntitiesRouter) create(req *http.Request) (interface{}, error) {
	entity := types.Entity{}
	if err := UnmarshalBody(req, &entity); err != nil {
		return nil, err
	}
	err := r.controller.Create(req.Context(), entity)
	return entity, err
}

func (r *EntitiesRouter) createOrReplace(req *http.Request) (interface{}, error) {
	entity := types.Entity{}
	if err := UnmarshalBody(req, &entity); err != nil {
		return nil, err
	}

	return entity, r.controller.CreateOrReplace(req.Context(), entity)
}
