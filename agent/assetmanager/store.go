package assetmanager

import (
	"sort"
	"strings"
	"sync"

	"github.com/sensu/sensu-go/types"
)

// A AssetStore holds state of runtime assets
type AssetStore struct {
	assets    map[string]*RuntimeAsset
	assetSets map[string]*RuntimeAssetSet
	rwMutex   *sync.RWMutex
}

type newAssetFn func(*types.Asset) *RuntimeAsset
type newSetFn func([]*RuntimeAsset) *RuntimeAssetSet

// NewAssetStore ...
func NewAssetStore() *AssetStore {
	return &AssetStore{
		assets:    make(map[string]*RuntimeAsset),
		assetSets: make(map[string]*RuntimeAssetSet),
		rwMutex:   &sync.RWMutex{},
	}
}

// FetchAsset - fetches asset from store, otherwise creates & adds it
func (storePtr *AssetStore) FetchAsset(asset *types.Asset, newFn newAssetFn) *RuntimeAsset {
	key := asset.Sha512

	// Return asset if it is already in the store
	if runtimeAsset := storePtr.getAsset(key); runtimeAsset != nil {
		return runtimeAsset
	}

	// Instantiate & store
	runtimeAsset := newFn(asset)
	storePtr.setAsset(key, runtimeAsset)

	return runtimeAsset
}

func (storePtr *AssetStore) getAsset(key string) *RuntimeAsset {
	storePtr.rwMutex.RLock()
	defer storePtr.rwMutex.RUnlock()
	return storePtr.assets[key]
}

func (storePtr *AssetStore) setAsset(key string, asset *RuntimeAsset) {
	storePtr.rwMutex.Lock()
	defer storePtr.rwMutex.Unlock()
	storePtr.assets[key] = asset
}

// FetchSet - fetches set from store, otherwise creatres & adds it
func (storePtr *AssetStore) FetchSet(assets []*RuntimeAsset, newFn newSetFn) *RuntimeAssetSet {
	key := concatAssetSetKey(assets)

	// Return set if it is already stored
	if set := storePtr.getAssetSet(key); set != nil {
		return set
	}

	// Instantiate & store
	set := newFn(assets)
	storePtr.setAssetSet(key, set)

	return set
}

func (storePtr *AssetStore) getAssetSet(key string) *RuntimeAssetSet {
	storePtr.rwMutex.RLock()
	defer storePtr.rwMutex.RUnlock()
	return storePtr.assetSets[key]
}

func (storePtr *AssetStore) setAssetSet(key string, set *RuntimeAssetSet) {
	storePtr.rwMutex.Lock()
	defer storePtr.rwMutex.Unlock()
	storePtr.assetSets[key] = set
}

// Clear - clears store's state
func (storePtr *AssetStore) Clear() {
	storePtr.rwMutex.Lock()
	storePtr.assets = make(map[string]*RuntimeAsset)
	storePtr.assetSets = make(map[string]*RuntimeAssetSet)
	storePtr.rwMutex.Unlock()
}

// Generate unique key for assets sets; concats first chars of each asset hash
func concatAssetSetKey(runtimeAssets []*RuntimeAsset) string {
	names := make([]string, len(runtimeAssets))
	for _, runtimeAsset := range runtimeAssets {
		names = append(names, runtimeAsset.asset.Sha512[:7])
	}

	sort.Strings(names)
	return strings.Join(names, "")
}
