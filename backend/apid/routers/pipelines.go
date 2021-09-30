package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/store"
)

// PipelinesRouter handles requests for /pipelines
type PipelinesRouter struct {
	handlers handlers.Handlers
}

// NewPipelineRouter instantiates new router for controlling pipeline resources
func NewPipelinesRouter(store store.ResourceStore) *PipelinesRouter {
	return &PipelinesRouter{
		handlers: handlers.Handlers{
			Resource: &corev2.Pipeline{},
			Store:    store,
		},
	}
}

// Mount the PipelinesRouter to a parent Router
func (r *PipelinesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:pipelines}",
	}

	routes.Get(r.handlers.GetResource)
	routes.List(r.handlers.ListResources, corev2.PipelineFields)
	routes.ListAllNamespaces(r.handlers.ListResources, "/{resource:pipelines}", corev2.PipelineFields)
	routes.Patch(r.handlers.PatchResource)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
	routes.Del(r.handlers.DeleteResource)
}
