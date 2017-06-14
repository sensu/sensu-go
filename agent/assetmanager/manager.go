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
}

// New - given agent returns instantiated Manager
func New(agentCacheDir string) *Manager {
	manager := &Manager{}
	manager.store = NewAssetStore()
	manager.factory = &AssetFactory{
		CacheDir: agentCacheDir,
		BaseEnv:  getSystemEnviron(),
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
	mngrPtr.factory.BaseEnv = getSystemEnviron()
	mngrPtr.store.Clear()
}

// Get system ENV variables and append any PATH, LD_LIBRARY_PATH,  & CPATH if
// missing.
func getSystemEnviron() []string {
	env := os.Environ()
	presentVars := map[string]bool{
		"PATH":            false,
		"LD_LIBRARY_PATH": false,
		"CPATH":           false,
	}

	for _, e := range env {
		pair := strings.Split(e, "=")
		key, _ := pair[0], pair[1]

		if presentVars[key] != true {
			presentVars[key] = true
		}
	}

	for key, val := range presentVars {
		if val == false {
			env = append(env, key+"=")
		}
	}

	return env
}
