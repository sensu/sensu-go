package routers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// AssetsRouter handles requests for /assets
type AssetsRouter struct {
	controller actions.AssetController
}

// NewAssetRouter instantiates new router for controlling asset resources
func NewAssetRouter(store store.AssetStore) *AssetsRouter {
	return &AssetsRouter{
		controller: actions.NewAssetController(store),
	}
}

// Mount the AssetsRouter to a parent Router
func (r *AssetsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:assets}",
	}

	routes.Get(r.find)
	routes.List(r.controller.List, corev2.AssetFields)
	routes.ListAllNamespaces(r.controller.List, "/{resource:assets}", corev2.AssetFields)
	routes.Post(r.create)
	routes.Put(r.createOrReplace)
}

func (r *AssetsRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	assetPath, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	record, err := r.controller.Find(req.Context(), assetPath)
	return record, err
}

func (r *AssetsRouter) create(req *http.Request) (interface{}, error) {
	cfg := types.Asset{}
	if err := UnmarshalBody(req, &cfg); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), cfg)
	return cfg, err
}

func (r *AssetsRouter) createOrReplace(req *http.Request) (interface{}, error) {
	var asset types.Asset
	if err := UnmarshalBody(req, &asset); err != nil {
		return nil, err
	}
	err := r.controller.CreateOrReplace(req.Context(), asset)
	return asset, err
}
