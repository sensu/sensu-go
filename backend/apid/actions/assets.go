package actions

import (
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

// AssetController expose actions in which a viewer can perform.
type AssetController struct {
	Store  store.AssetStore
	Policy authorization.AssetPolicy
}

// NewAssetController returns new AssetController
func NewAssetController(store store.AssetStore) AssetController {
	return AssetController{
		Store:  store,
		Policy: authorization.Assets,
	}
}

// Query returns resources available to the viewer filter by given params.
func (a AssetController) Query(ctx context.Context, params QueryParams) ([]interface{}, error) {
	abilities := a.Policy.WithContext(ctx)

	// Fetch from store
	results, serr := a.Store.GetAssets(ctx)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	// Filter out those resources the viewer does not have access to view.
	resources := []interface{}{}
	for _, result := range results {
		if yes := abilities.CanRead(result); yes {
			resources = append(resources, result)
		}
	}

	return resources, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a AssetController) Find(ctx context.Context, params QueryParams) (interface{}, error) {
	// Validate params
	if id := params["id"]; id == "" {
		return nil, NewErrorf(InternalErr, "'id' param missing")
	}

	// Fetch from store
	result, serr := a.Store.GetAssetByName(ctx, params["id"])
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	// Verify user has permission to view
	abilities := a.Policy.WithContext(ctx)
	if result != nil && abilities.CanRead(result) {
		return result, nil
	}

	return nil, NewErrorf(NotFound)
}

// Create instatiates, validates and persists new resource if viewer has access.
func (a AssetController) Create(ctx context.Context, newAsset types.Asset) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &newAsset)
	abilities := a.Policy.WithContext(ctx)

	// Check for existing
	if e, err := a.Store.GetAssetByName(ctx, newAsset.Name); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Verify viewer can make change
	if yes := abilities.CanCreate(); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Validate
	if err := newAsset.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.Store.UpdateAsset(ctx, &newAsset); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}
