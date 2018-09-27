package assetmanager

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sensu/sensu-go/types"
)

// Manager manages caching & installation of dependencies
type Manager struct {
	factory *AssetFactory
	store   *AssetStore
	entity  *types.Entity
}

// New - given agent returns instantiated Manager
func New(agentCacheDir string, entity *types.Entity) *Manager {
	manager := &Manager{}
	manager.entity = entity
	manager.store = NewAssetStore()
	manager.factory = &AssetFactory{
		CacheDir: agentCacheDir,
		BaseEnv:  getSystemEnviron(os.Environ()),
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

	filteredRuntimeAssets := []*RuntimeAsset{}
	for _, runtimeAsset := range runtimeAssets {
		if relevant, err := runtimeAsset.isRelevantTo(*mngrPtr.entity); err != nil {
			logger.Debugf("asset '%s' was filtered", runtimeAsset.asset.Name)
		} else if !relevant {
			logger.Debugf("asset '%s' was filtered", runtimeAsset.asset.Name)
		} else {
			filteredRuntimeAssets = append(filteredRuntimeAssets, runtimeAsset)
		}
	}

	return mngrPtr.store.FetchSet(filteredRuntimeAssets, mngrPtr.factory.NewAssetSet)
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
	mngrPtr.factory.BaseEnv = getSystemEnviron(os.Environ())
	mngrPtr.store.Clear()
}

// Get system ENV variables and append any PATH, LD_LIBRARY_PATH,  & CPATH if
// missing.
func getSystemEnviron(env []string) []string {
	presentVars := map[string]bool{
		"PATH":            false,
		"LD_LIBRARY_PATH": false,
		"CPATH":           false,
	}

	for _, e := range env {
		pair := strings.Split(e, "=")
		// Transform the key to uppercase because Windows uses "Path" and not "PATH"
		key := strings.ToUpper(pair[0])

		// Mark it as present
		presentVars[key] = true
	}

	for key, val := range presentVars {
		if !val {
			env = append(env, key+"=")
		}
	}

	return env
}
