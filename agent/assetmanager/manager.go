package assetmanager

import (
	"os"
	"path/filepath"

	"github.com/sensu/sensu-go/types"
)

// Manager manages caching & installation of dependencies
type Manager struct {
	factory *AssetFactory
	store   *AssetStore
}

// New - given agent returns instantiated Manager
func New(agentCacheDir string) *Manager {
	manager := &Manager{}
	manager.store = NewAssetStore()
	manager.factory = &AssetFactory{
		CacheDir: agentCacheDir,
		BaseEnv:  os.Environ(),
	}

	return manager
}

// RegisterSet - registers given assets and returns resulting set
func (mngrPtr *Manager) RegisterSet(assets []types.Asset) *RuntimeAssetSet {
	runtimeAssets := make([]*RuntimeAsset, len(assets))
	for i, asset := range assets {
		runtimeAsset := mngrPtr.store.FetchAsset(&asset, mngrPtr.factory.NewAsset)
		runtimeAssets[i] = runtimeAsset
	}

	return mngrPtr.store.FetchSet(runtimeAssets, mngrPtr.factory.NewAssetSet)
}

// SetCacheDir sets cache directory given a base directory
func (mngrPtr *Manager) SetCacheDir(baseDir string) {
	mngrPtr.factory.CacheDir = filepath.Join(baseDir, depsCachePath)
	mngrPtr.store.Clear()
}

// Reset clears all knownAssets and env from state, this forces the agent to
// recompute the next time a check is run.
//
// NOTE: Cache on disk is not cleared.
func (mngrPtr *Manager) Reset() {
	mngrPtr.factory.BaseEnv = os.Environ()
	mngrPtr.store.Clear()
}
