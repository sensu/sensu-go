package api

import (
	"context"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"

	corev2 "github.com/sensu/core/v2"
)

// AssetClient is an API client for assets.
type AssetClient struct {
	client GenericClient
	auth   authorization.Authorizer
}

// NewAssetClient creates a new AssetClient, given a store and an authorizer.
func NewAssetClient(store storev2.Interface, auth authorization.Authorizer) *AssetClient {
	return &AssetClient{
		client: GenericClient{
			Store:      store,
			Auth:       auth,
			Kind:       &corev2.Asset{},
			APIGroup:   "core",
			APIVersion: "v2",
		},
		auth: auth,
	}
}

// ListAssets fetches a list of asset resources, if authorized.
func (a *AssetClient) ListAssets(ctx context.Context) ([]*corev2.Asset, error) {
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.Asset{}
	if err := a.client.List(ctx, &slice, pred); err != nil {
		return nil, err
	}
	return slice, nil
}

// FetchAsset fetches an asset resource from the backend, if authorized.
func (a *AssetClient) FetchAsset(ctx context.Context, name string) (*corev2.Asset, error) {
	var asset corev2.Asset
	if err := a.client.Get(ctx, name, &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

// CreateAsset creates an asset resource, if authorized.
func (a *AssetClient) CreateAsset(ctx context.Context, asset *corev2.Asset) error {
	if err := a.client.Create(ctx, asset); err != nil {
		return err
	}
	return nil
}

// UpdateAsset updates an asset resource, if authorized.
func (a *AssetClient) UpdateAsset(ctx context.Context, asset *corev2.Asset) error {
	if err := a.client.Update(ctx, asset); err != nil {
		return err
	}
	return nil
}
