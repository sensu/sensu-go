package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/store"
)

// EntitiesRouter handles requests for /entities
type EntitiesRouter struct {
	handlers handlers.Handlers
}

// NewEntitiesRouter instantiates new router for controlling entities resources
func NewEntitiesRouter(store store.ResourceStore) *EntitiesRouter {
	return &EntitiesRouter{
		handlers: handlers.Handlers{
			Resource: &corev2.Entity{},
			Store:    store,
		},
	}
}

// Mount the EntitiesRouter to a parent Router
func (r *EntitiesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:entities}",
	}

	routes.Del(r.handlers.DeleteResource)
	routes.Get(r.handlers.GetResource)
	routes.List(r.handlers.ListResources, corev2.EntityFields)
	routes.ListAllNamespaces(r.handlers.ListResources, "/{resource:entities}", corev2.EntityFields)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
}
