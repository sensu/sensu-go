package asset

import (
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/eval"
)

// NewFilteredManager returns an asset Getter that filters assets based on the
// given entity. Assets that aren't filtered get passed to the underlying
// getter, allowing composition with other asset managers.
func NewFilteredManager(getter Getter, entity *types.Entity) *filteredManager {
	return &filteredManager{
		getter: getter,
		entity: entity,
	}
}

// filteredManager evaluates asset filter(s) and calls the underlying
// asset Getter based on filter result.
type filteredManager struct {
	entity *types.Entity
	getter Getter
}

// Get fetches, verifies, and expands an asset, but only if it is not
// filtered.
func (f *filteredManager) Get(asset *types.Asset) (*RuntimeAsset, error) {
	filtered, err := f.isFiltered(asset)
	if err != nil {
		return nil, err
	}

	if filtered {
		return nil, nil
	}

	return f.getter.Get(asset)
}

// isFiltered evaluates the given asset's filters and returns true if all of
// them match the current entity.
func (f *filteredManager) isFiltered(asset *types.Asset) (bool, error) {
	if len(asset.Filters) == 0 {
		return false, nil
	}

	params := make(map[string]interface{}, 1)
	params["entity"] = f.entity

	for _, filter := range asset.Filters {
		result, err := eval.EvaluatePredicate(filter, params)
		if err != nil {
			switch err.(type) {
			case eval.SyntaxError, eval.TypeError:
				return result, err
			default:
				// Other errors during execution are likely due to missing
				// attrs, simply continue in this case.
				continue
			}
		}

		if !result {
			return result, err
		}
	}

	return true, nil
}
