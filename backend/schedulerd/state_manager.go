package schedulerd

import (
	"strings"
	"sync"

	"github.com/sensu/sensu-go/types"
)

// A StateManager keeps copies of unmarshal'd resources schedulerd requires to run
// efficiently schedule checks.
type StateManager struct {
	state *SchedulerState
	mutex *sync.Mutex
}

// NewStateManager returns a new instance of schedulerd's cache
func NewStateManager() *StateManager {
	return &StateManager{
		state: &SchedulerState{},
		mutex: &sync.Mutex{},
	}
}

// State returns reference to current state of the cache
func (cachePtr *StateManager) State() *SchedulerState {
	return cachePtr.state
}

// Update synchronously updates state w/ result of closure
func (cachePtr *StateManager) Update(updateFn func(newState *SchedulerState)) {
	// Shallow copy contents
	newState := SchedulerState{
		checks: cachePtr.state.checks,
		assets: cachePtr.state.assets,
	}

	// Lock to avoid competing updates
	cachePtr.mutex.Lock()
	defer cachePtr.mutex.Unlock()

	// Pass to caller
	updateFn(&newState)
	cachePtr.state = &newState
}

// A SchedulerState represents the internal state of the cache
type SchedulerState struct {
	checks map[string]*types.CheckConfig
	assets map[string]map[string]*types.Asset
}

// GetCheck returns check given name and organization
func (statePtr *SchedulerState) GetCheck(name, org string) *types.CheckConfig {
	key := concatUniqueKey(name, org)
	return statePtr.checks[key]
}

// GetAssetsInOrg returns all assets associated given organization
func (statePtr *SchedulerState) GetAssetsInOrg(org string) (res []*types.Asset) {
	for _, asset := range statePtr.assets[org] {
		res = append(res, asset)
	}
	return
}

func (statePtr *SchedulerState) SetChecks(checks []*types.CheckConfig) {
	statePtr.checks = make(map[string]*types.CheckConfig)
	for _, check := range checks {
		statePtr.addCheck(check)
	}
}

func (statePtr *SchedulerState) SetAssets(assets []*types.Asset) {
	statePtr.assets = make(map[string]map[string]*types.Asset)
	for _, asset := range assets {
		statePtr.addAsset(asset)
	}
}

func (statePtr *SchedulerState) addCheck(check *types.CheckConfig) {
	key := concatUniqueKey(check.Name, check.Organization)
	statePtr.checks[key] = check
}

func (statePtr *SchedulerState) addAsset(asset *types.Asset) {
	org := asset.Organization
	if orgMap := statePtr.assets[org]; orgMap == nil {
		statePtr.assets[org] = make(map[string]*types.Asset)
	}

	key := concatUniqueKey(asset.Name, org)
	statePtr.assets[org][key] = asset
}

func concatUniqueKey(name, org string) string {
	return strings.Join([]string{name, org}, "-")
}
