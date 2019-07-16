package api

import (
	"context"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

type AssetClient struct {
	store    store.ResourceStore
	auth     authorization.Authorizer
	resource corev2.Resource
}

func NewAssetClient(store store.ResourceStore, auth authorization.Authorizer) *AssetClient {
	return &AssetClient{
		store:    store,
		auth:     auth,
		resource: &corev2.Asset{},
	}
}

// ListAssets fetches a list of asset resources
func (a *AssetClient) ListAssets(ctx context.Context) ([]*corev2.Asset, error) {
	attrs := assetListAttributes(ctx)
	if err := authorize(ctx, a.auth, attrs); err != nil {
		return nil, err
	}
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.Asset{}
	if err := a.store.ListResources(ctx, a.resource.StorePrefix(), &slice, pred); err != nil {
		return nil, err
	}
	return slice, nil
}

// FetchAsset fetches an asset resource from the backend
func (a *AssetClient) FetchAsset(ctx context.Context, name string) (*types.Asset, error) {
	attrs := assetFetchAttributes(ctx, name)
	if err := authorize(ctx, a.auth, attrs); err != nil {
		return nil, err
	}
	var asset corev2.Asset
	return &asset, a.store.GetResource(ctx, name, &asset)
}

// CreateAsset creates an asset resource
func (a *AssetClient) CreateAsset(ctx context.Context, asset *corev2.Asset) error {
	attrs := assetCreateAttributes(ctx, asset.Name)
	if err := authorize(ctx, a.auth, attrs); err != nil {
		return err
	}
	return a.store.CreateResource(ctx, asset)
}

// UpdateAsset updates an asset resource
func (a *AssetClient) UpdateAsset(ctx context.Context, asset *corev2.Asset) error {
	attrs := assetUpdateAttributes(ctx, asset.Name)
	if err := authorize(ctx, a.auth, attrs); err != nil {
		return err
	}
	return a.store.CreateOrUpdateResource(ctx, asset)
}

func assetListAttributes(ctx context.Context) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:   "core",
		APIVersion: "v2",
		Namespace:  corev2.ContextNamespace(ctx),
		Resource:   "assets",
		Verb:       "list",
	}
}

func assetFetchAttributes(ctx context.Context, name string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     "assets",
		Verb:         "get",
		ResourceName: name,
	}
}

func assetCreateAttributes(ctx context.Context, name string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     "assets",
		Verb:         "create",
		ResourceName: name,
	}
}

func assetUpdateAttributes(ctx context.Context, name string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     "assets",
		Verb:         "update",
		ResourceName: name,
	}
}
