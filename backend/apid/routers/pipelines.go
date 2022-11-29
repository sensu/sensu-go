package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// PipelinesRouter handles requests for /pipelines
type PipelinesRouter struct {
	store storev2.Interface
}

// NewPipelineRouter instantiates new router for controlling pipeline resources
func NewPipelinesRouter(store storev2.Interface) *PipelinesRouter {
	return &PipelinesRouter{
		store: store,
	}
}

// Mount the PipelinesRouter to a parent Router
func (r *PipelinesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:pipelines}",
	}

	handlers := handlers.NewHandlers[*corev2.Pipeline](r.store)

	routes.Get(handlers.GetResource)
	routes.List(handlers.ListResources, corev2.PipelineFields)
	routes.ListAllNamespaces(handlers.ListResources, "/{resource:pipelines}", corev2.PipelineFields)
	routes.Patch(handlers.PatchResource)
	routes.Post(handlers.CreateResource)
	routes.Put(handlers.CreateOrUpdateResource)
	routes.Del(handlers.DeleteResource)
}
