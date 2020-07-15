package routers

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
)

// EntitiesRouter handles requests for /entities
type EntitiesRouter struct {
	controller EntityController
	store      store.Store
	eventStore store.EventStore
}

type EntityController interface {
	Find(ctx context.Context, id string) (*corev2.Entity, error)
	List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error)
	Create(ctx context.Context, entity corev2.Entity) error
	CreateOrReplace(ctx context.Context, entity corev2.Entity) error
}

// NewEntitiesRouter instantiates new router for controlling entities resources
func NewEntitiesRouter(store store.Store, events store.EventStore) *EntitiesRouter {
	return &EntitiesRouter{
		controller: actions.NewEntityController(store),
		store:      store,
		eventStore: events,
	}
}

// Mount the EntitiesRouter to a parent Router
func (r *EntitiesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:entities}",
	}

	deleter := actions.EntityDeleter{
		EntityStore: r.store,
		EventStore:  r.eventStore,
	}

	routes.Del(deleter.Delete)
	routes.Get(r.find)
	routes.List(r.controller.List, corev2.EntityFields)
	routes.ListAllNamespaces(r.controller.List, "/{resource:entities}", corev2.EntityFields)
	routes.Post(r.create)
	routes.Put(r.createOrReplace)
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

func (r *EntitiesRouter) create(req *http.Request) (interface{}, error) {
	entity := corev2.Entity{}
	if err := UnmarshalBody(req, &entity); err != nil {
		return nil, err
	}
	err := r.controller.Create(req.Context(), entity)
	return entity, err
}

func (r *EntitiesRouter) createOrReplace(req *http.Request) (interface{}, error) {
	entity := corev2.Entity{}
	if err := UnmarshalBody(req, &entity); err != nil {
		return nil, err
	}

	return entity, r.controller.CreateOrReplace(req.Context(), entity)
}
