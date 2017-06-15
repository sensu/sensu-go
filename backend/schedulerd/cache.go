package schedulerd

import (
	"strings"

	"github.com/sensu/sensu-go/types"
)

// A Cache keeps copies of unmarshal'd resources schedulerd requires to run
// efficiently schedule checks.
type Cache struct {
	state *CacheState
	mutex *sync.Mutex
}

// NewCache returns a new instance of schedulerd's cache
func NewCache() *Cache {
	return &Cache{
		state: &CacheState{},
		mutex: &sync.Mutex{},
	}
}

// SetChecks synchronously updates the checks in our state
func (cachePtr *CheckCache) SetChecks(checks []*types.CheckConfig) {
	cachePtr.update(func(newState *CacheState) {
		newState.setChecks(checks)
	})
}

// SetAssets synchronously updates the assets in our state
func (cachePtr *CheckCache) SetAssets(assets []*types.Asset) {
	cachePtr.update(func(newState *CacheState) {
		newState.setAssets(assets)
	})
}

// State returns reference to current state of the cache
func (cachePtr *CheckCache) State() *CacheState {
	return cachePtr.state
}

func (cachePtr *Cache) update(updateFn func(newState *CacheState)) {
	var newState *CacheState

	// Lock to avoid competing updates
	cachePtr.mutex.Lock()
	defer cachePtr.mutex.Unlock()

	// Copy & Pass to caller
	*newState = *cachePtr.cache
	updateFn(newState)
}

// A CacheState represents the internal state of the cache
type CacheState struct {
	checks map[string]*types.Check
	assets map[string]*types.Asset
}

// GetCheck returns check given name and organization
func (statePtr *CacheState) GetCheck(name, org string) *types.Check {
	key := concatUniqueKey(name, org)
	return storePtr.checks[key]
}

// GetAssetsInOrg returns all assets associated given organization
func (statePtr *CacheState) GetAssetsInOrg(org string) (res []*types.Asset) {
	for _, asset := range storePtr.assets {
		if org == asset.Organization {
			res = append(res, asset)
		}
	}
	return
}

func (statePtr *CacheState) setChecks(checks []*types.Check) {
	statePtr.checks = make(map[string]*types.Check)
	for _, check := range statePtr.checks {
		statePtr.addCheck(check)
	}
}

func (statePtr *CacheState) setAssets(assets []*types.Asset) {
	statePtr.assets = make(map[string]*types.Asset)
	for _, asset := range statePtr.assets {
		statePtr.addAsset(asset)
	}
}

func (statePtr *CacheState) addCheck(check *types.Check) {
	key := concatUniqueKey(check.Name, check.Organization)
	storePtr.checks[key] = check
}

func (statePtr *CacheState) addAsset(asset *types.Asset) {
	key := concatUniqueKey(asset.Name, asset.Organization)
	assetPtr.assets[key] = asset
}

func concatUniqueKey(name, org string) string {
	return strings.Join([]string{name, org}, "-")
}
