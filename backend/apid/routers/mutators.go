package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// MutatorsRouter handles /mutators requests.
type MutatorsRouter struct {
	store storev2.Interface
}

// NewMutatorsRouter creates a new MutatorsRouter.
func NewMutatorsRouter(store storev2.Interface) *MutatorsRouter {
	return &MutatorsRouter{
		store: store,
	}
}

// Mount the MutatorsRouter to a parent Router
func (r *MutatorsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:mutators}",
	}

	handlers := handlers.NewHandlers[*corev2.Mutator](r.store)

	routes.Del(handlers.DeleteResource)
	routes.Get(handlers.GetResource)
	routes.List(handlers.ListResources, corev3.MutatorFields)
	routes.ListAllNamespaces(handlers.ListResources, "/{resource:mutators}", corev3.MutatorFields)
	routes.Patch(handlers.PatchResource)
	routes.Post(handlers.CreateResource)
	routes.Put(handlers.CreateOrUpdateResource)
}
