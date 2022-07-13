package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// AssetsRouter handles requests for /assets
type AssetsRouter struct {
	handlers handlers.Handlers
}

// NewAssetRouter instantiates new router for controlling asset resources
func NewAssetRouter(store storev2.Interface) *AssetsRouter {
	return &AssetsRouter{
		handlers: handlers.Handlers{
			Resource: &corev2.Asset{},
			Store:    store,
		},
	}
}

// Mount the AssetsRouter to a parent Router
func (r *AssetsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:assets}",
	}

	routes.Get(r.handlers.GetResource)
	routes.List(r.handlers.ListResources, corev2.AssetFields)
	routes.ListAllNamespaces(r.handlers.ListResources, "/{resource:assets}", corev2.AssetFields)
	routes.Patch(r.handlers.PatchResource)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
	routes.Del(r.handlers.DeleteResource)
}
