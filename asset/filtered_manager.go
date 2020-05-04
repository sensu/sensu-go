package asset

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/token"
	"github.com/sensu/sensu-go/types/dynamic"
	"github.com/sirupsen/logrus"
)

// NewFilteredManager returns an asset Getter that filters assets based on the
// given entity. Assets that aren't filtered get passed to the underlying
// getter, allowing composition with other asset managers.
func NewFilteredManager(getter Getter, entity *corev2.Entity) *filteredManager {
	return &filteredManager{
		getter: getter,
		entity: entity,
	}
}

// filteredManager evaluates asset filter(s) and calls the underlying
// asset Getter based on filter result.
type filteredManager struct {
	entity *corev2.Entity
	getter Getter
}

// Get fetches, verifies, and expands an asset, but only if it is filtered.
func (f *filteredManager) Get(ctx context.Context, asset *corev2.Asset) (*RuntimeAsset, error) {
	var filteredAsset *corev2.Asset

	fields := logrus.Fields{
		"entity":  f.entity.Name,
		"asset":   asset.Name,
		"filters": asset.Filters,
	}

	if len(asset.Builds) > 0 {
		fields = logrus.Fields{
			"entity": f.entity.Name,
			"asset":  asset.Name,
		}
		logger.WithFields(fields).Info("asset includes builds, using builds instead of asset")
		for _, build := range asset.Builds {
			assetBuild := &corev2.Asset{
				URL:        build.URL,
				Sha512:     build.Sha512,
				Filters:    build.Filters,
				Headers:    build.Headers,
				ObjectMeta: asset.ObjectMeta,
			}

			buildFields := logrus.Fields{
				"filter": assetBuild.Filters,
			}

			filtered, err := f.isFiltered(assetBuild)
			if err != nil {
				logger.WithFields(fields).WithFields(buildFields).WithError(err).Error("error filtering entities from asset build")
				return nil, err
			}

			if !filtered {
				logger.WithFields(fields).WithFields(buildFields).Debug("entity not filtered, not installing asset build")
				continue
			}

			logger.WithFields(fields).WithFields(buildFields).Debug("entity filtered, installing asset build")
			filteredAsset = assetBuild
			break
		}
	} else {
		filtered, err := f.isFiltered(asset)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("error filtering entities from asset")
			return nil, err
		}

		if !filtered {
			logger.WithFields(fields).Debug("entity not filtered, not installing asset")
			return nil, nil
		}

		logger.WithFields(fields).Debug("entity filtered, installing asset")
		filteredAsset = asset
	}

	// catch case where no asset build filters pass
	if filteredAsset == nil {
		logger.WithFields(fields).Warn("entity not filtered from any asset builds, not installing asset")
		return nil, nil
	}

	// Perform token substitution on the asset before retrieving it
	if err := token.SubstituteAsset(filteredAsset, f.entity); err != nil {
		logger.WithField("entity", f.entity).Debug(err)
		return nil, fmt.Errorf("error while substituting asset %q tokens: %s", asset.Name, err)
	}

	return f.getter.Get(ctx, filteredAsset)
}

// isFiltered evaluates the given asset's filters and returns true if all of
// them match the current entity.
func (f *filteredManager) isFiltered(asset *corev2.Asset) (bool, error) {
	if len(asset.Filters) == 0 {
		return true, nil
	}

	synth := dynamic.Synthesize(f.entity)
	params := map[string]interface{}{"entity": synth}

	for _, filter := range asset.Filters {
		result, err := js.Evaluate(filter, params, nil)
		if err != nil || !result {
			return false, err
		}
	}

	return true, nil
}
