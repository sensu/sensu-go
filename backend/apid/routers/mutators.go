package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/store"
)

// MutatorsRouter handles /mutators requests.
type MutatorsRouter struct {
	handlers handlers.Handlers
}

// NewMutatorsRouter creates a new MutatorsRouter.
func NewMutatorsRouter(store store.ResourceStore) *MutatorsRouter {
	return &MutatorsRouter{
		handlers: handlers.Handlers{
			Resource: &corev2.Mutator{},
			Store:    store,
		},
	}
}

// Mount the MutatorsRouter to a parent Router
func (r *MutatorsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:mutators}",
	}

	routes.Del(r.handlers.DeleteResource)
	routes.Get(r.handlers.GetResource)
	routes.List(r.handlers.ListResources, corev2.MutatorFields)
	routes.ListAllNamespaces(r.handlers.ListResources, "/{resource:mutators}", corev2.MutatorFields)
	routes.Patch(r.handlers.PatchResource)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
}
