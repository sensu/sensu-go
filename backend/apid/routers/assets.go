package routers

import (
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// AssetsRouter handles requests for /assets
type AssetsRouter struct {
	store storev2.Interface
}

// NewAssetRouter instantiates new router for controlling asset resources
func NewAssetRouter(store storev2.Interface) *AssetsRouter {
	return &AssetsRouter{
		store: store,
	}
}

// Mount the AssetsRouter to a parent Router
func (r *AssetsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:assets}",
	}

	handlers := handlers.NewHandlers[*corev2.Asset](r.store)

	routes.Get(handlers.GetResource)
	routes.List(handlers.ListResources, corev3.AssetFields)
	routes.ListAllNamespaces(handlers.ListResources, "/{resource:assets}", corev3.AssetFields)
	routes.Patch(handlers.PatchResource)
	routes.Post(handlers.CreateResource)
	routes.Put(handlers.CreateOrUpdateResource)
	routes.Del(handlers.DeleteResource)
}
