package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/store"
)

// MANISHA PIPELINE ROUTING HERE
// PipelinesRouter handles requests for /pipelines
type FallbackPipelinesRouter struct {
	handlers handlers.Handlers
}

// NewPipelineRouter instantiates new router for controlling pipeline resources
func NewFallbackPipelinesRouter(store store.ResourceStore) *FallbackPipelinesRouter {
	return &FallbackPipelinesRouter{
		handlers: handlers.Handlers{
			Resource: &corev2.FallbackPipeline{},
			Store:    store,
		},
	}
}

// Mount the PipelinesRouter to a parent Router
func (r *FallbackPipelinesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:fallbackPipeline}",
	}

	routes.Get(r.handlers.GetResource)
	routes.List(r.handlers.ListResources, corev2.FallbackPipelineFields)
	routes.ListAllNamespaces(r.handlers.ListResources, "/{resource:fallbackPipeline}", corev2.FallbackPipelineFields)
	routes.Patch(r.handlers.PatchResource)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
	routes.Del(r.handlers.DeleteResource)
}
