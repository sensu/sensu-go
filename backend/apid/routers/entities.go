package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/store"
)

// EntitiesRouter handles requests for /entities
type EntitiesRouter struct {
	handlers   handlers.Handlers
	store      store.Store
	eventStore store.EventStore
}

// NewEntitiesRouter instantiates new router for controlling entities resources
func NewEntitiesRouter(store store.Store, events store.EventStore) *EntitiesRouter {
	return &EntitiesRouter{
		handlers: handlers.Handlers{
			Resource: &corev2.Entity{},
			Store:    store,
		},
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
	routes.Get(r.handlers.GetResource)
	routes.List(r.handlers.ListResources, corev2.EntityFields)
	routes.ListAllNamespaces(r.handlers.ListResources, "/{resource:entities}", corev2.EntityFields)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
}
