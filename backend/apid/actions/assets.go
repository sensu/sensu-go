package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// assetUpdateFields whitelists fields allowed to be updated for Assets
var assetUpdateFields = []string{
	"Sha512",
	"URL",
}

// AssetController expose actions in which a viewer can perform.
type AssetController struct {
	store store.AssetStore
}

// NewAssetController returns new AssetController
func NewAssetController(store store.AssetStore) AssetController {
	return AssetController{
		store: store,
	}
}

// List returns resources available to the viewer filter by given params.
func (a AssetController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	// Fetch from store
	results, err := a.store.GetAssets(ctx, pred)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	resources := make([]corev2.Resource, len(results))
	for i, v := range results {
		resources[i] = corev2.Resource(v)
	}

	return resources, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a AssetController) Find(ctx context.Context, name string) (*types.Asset, error) {
	// Validate params
	if id := name; id == "" {
		return nil, NewErrorf(InternalErr, "'id' param missing")
	}

	// Fetch from store
	result, serr := a.store.GetAssetByName(ctx, name)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}
	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

// Create instatiates, validates and persists new resource if viewer has access.
func (a AssetController) Create(ctx context.Context, newAsset types.Asset) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &newAsset)

	// Check for existing
	if e, err := a.store.GetAssetByName(ctx, newAsset.Name); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Validate
	if err := newAsset.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.store.UpdateAsset(ctx, &newAsset); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// CreateOrReplace creates or replaces the asset given.
func (a AssetController) CreateOrReplace(ctx context.Context, asset types.Asset) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &asset)

	// Validate
	if err := asset.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist Changes
	if serr := a.store.UpdateAsset(ctx, &asset); serr != nil {
		return NewError(InternalErr, serr)
	}

	return nil
}
