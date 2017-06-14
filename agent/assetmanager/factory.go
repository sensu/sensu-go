package assetmanager

import "github.com/sensu/sensu-go/types"

// AssetFactory helps instantiate new runtime assets & sets w/ given contextual
// details.
type AssetFactory struct {
	// BaseEnv is a set of ENV variables to apply when creating asset sets
	BaseEnv []string

	// CacheDir is the directory where assets are stored
	CacheDir string
}

// NewAsset returns a new RuntimeAsset given an asset
func (factory AssetFactory) NewAsset(asset *types.Asset) *RuntimeAsset {
	return NewRuntimeAsset(asset, factory.CacheDir)
}

// NewAssetSet returns a new RuntimeAsset given an asset
func (factory AssetFactory) NewAssetSet(assets []*RuntimeAsset) *RuntimeAssetSet {
	return NewRuntimeAssetSet(assets, factory.BaseEnv)
}
