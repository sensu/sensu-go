package actions

import (
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
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
