package routers

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/apid/request"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// EntitiesRouter handles requests for /entities
type EntitiesRouter struct {
	controller      EntityController
	store           storev2.Interface
	configSubrouter EntityConfigRouter
}

type EntityConfigRouter struct {
	store storev2.Interface
}

type EntityController interface {
	Find(ctx context.Context, id string) (*corev2.Entity, error)
	List(ctx context.Context, pred *store.SelectionPredicate) ([]corev3.Resource, error)
	Create(ctx context.Context, entity corev2.Entity) error
	CreateOrReplace(ctx context.Context, entity corev2.Entity) error
}

// NewEntitiesRouter instantiates new router for controlling entities resources
func NewEntitiesRouter(store storev2.Interface) *EntitiesRouter {
	return &EntitiesRouter{
		controller: actions.NewEntityController(store),
		store:      store,
		configSubrouter: EntityConfigRouter{
			store: store,
		},
	}
}

// Mount the EntitiesRouter to a parent Router
func (r *EntitiesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:entities}",
	}

	deleter := actions.EntityDeleter{
		EntityStore: r.store.GetEntityStore(),
		EventStore:  r.store.GetEventStore(),
	}

	ecHandlers := handlers.NewHandlers[*corev3.EntityConfig](r.store)

	routes.Del(deleter.Delete)
	routes.Get(r.find)
	routes.List(r.controller.List, corev3.EntityFields)
	routes.ListAllNamespaces(r.controller.List, "/{resource:entities}", corev3.EntityFields)
	routes.Patch(ecHandlers.PatchResource)
	routes.Post(r.create)
	routes.Put(r.createOrReplace)
}

func (r *EntitiesRouter) find(req *http.Request) (corev3.Resource, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	return r.controller.Find(req.Context(), id)
}

func (r *EntitiesRouter) create(req *http.Request) (corev3.Resource, error) {
	entity, err := request.Resource[*corev2.Entity](req)
	if err != nil {
		return nil, err
	}
	err = r.controller.Create(req.Context(), *entity)
	return entity, err
}

func (r *EntitiesRouter) createOrReplace(req *http.Request) (corev3.Resource, error) {
	entity, err := request.Resource[*corev2.Entity](req)
	if err != nil {
		return nil, err
	}

	if entity.Labels[corev2.ManagedByLabel] == "sensu-agent" {
		return nil, actions.NewError(actions.AlreadyExistsErr, errors.New("entity is managed by its agent"))
	}

	return entity, r.controller.CreateOrReplace(req.Context(), *entity)
}
