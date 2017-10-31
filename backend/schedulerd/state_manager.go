package schedulerd

import (
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// SynchronizeMinInterval minimum interval inwhich we poll the store for updates
const SynchronizeMinInterval = uint(10)

// A StateManager keeps copies of unmarshal'd resources schedulerd requires to run
// efficiently schedule checks.
type StateManager struct {
	OnChecksChange func(state *SchedulerState)

	state *atomic.Value
	mutex *sync.Mutex

	synchronizer *SynchronizeStateScheduler
}

// NewStateManager returns a new instance of schedulerd's cache
func NewStateManager(store store.Store) *StateManager {
	manager := &StateManager{
		OnChecksChange: func(state *SchedulerState) {},

		state: &atomic.Value{},
		mutex: &sync.Mutex{},
	}
	manager.Swap(&SchedulerState{})

	manager.synchronizer = NewSynchronizeStateScheduler(
		SynchronizeMinInterval,
		&SyncronizeChecks{
			Store:    store,
			OnUpdate: manager.updateChecks,
		},
		&SyncronizeAssets{
			Store:    store,
			OnUpdate: manager.updateAssets,
		},
	)

	return manager
}

// Start keeping state synchronized
func (mngrPtr *StateManager) Start() {
	mngrPtr.synchronizer.Start()
}

// Stop keeping state synchronized
func (mngrPtr *StateManager) Stop() error {
	return mngrPtr.synchronizer.Stop()
}

// State returns reference to current state of the cache
func (mngrPtr *StateManager) State() *SchedulerState {
	return mngrPtr.state.Load().(*SchedulerState)
}

// Update synchronously updates state w/ result of closure
func (mngrPtr *StateManager) Update(updateFn func(newState *SchedulerState)) {
	mngrPtr.updateState(updateFn)

	state := mngrPtr.State()
	mngrPtr.OnChecksChange(state)
}

// Swap state atom
func (mngrPtr *StateManager) Swap(state *SchedulerState) {
	mngrPtr.state.Store(state)
}

func (mngrPtr *StateManager) updateChecks(checks []*types.CheckConfig) {
	mngrPtr.updateState(func(state *SchedulerState) {
		state.SetChecks(checks)
		mngrPtr.OnChecksChange(state)
	})

	mngrPtr.updateSyncInterval()
}

func (mngrPtr *StateManager) updateAssets(assets []*types.Asset) {
	mngrPtr.updateState(func(state *SchedulerState) {
		state.SetAssets(assets)
	})
}

func (mngrPtr *StateManager) updateState(updateFn func(newState *SchedulerState)) {
	// Lock to avoid competing updates
	mngrPtr.mutex.Lock()
	defer mngrPtr.mutex.Unlock()

	// Shallow copy contents of state
	oldState := mngrPtr.State()
	newState := &SchedulerState{}
	*newState = *oldState

	// Pass to caller
	updateFn(newState)
	mngrPtr.Swap(newState)
}

func (mngrPtr *StateManager) updateSyncInterval() {
	state := mngrPtr.State()

	// Find min interval
	minInterval := SynchronizeMinInterval
	for _, check := range state.checks {
		checkInterval := uint(check.Interval)
		if checkInterval < minInterval {
			minInterval = checkInterval
		}
	}

	// Set updated interval
	mngrPtr.synchronizer.SetInterval(minInterval)
}

// A SchedulerState represents the internal state of the cache
type SchedulerState struct {
	checks map[string]*types.CheckConfig
	assets map[string]map[string]*types.Asset
}

// GetCheck returns check given name and organization
func (statePtr *SchedulerState) GetCheck(name, org, env string) *types.CheckConfig {
	key := concatUniqueKey(name, org, env)
	return statePtr.checks[key]
}

// GetAssetsInOrg returns all assets associated given organization
func (statePtr *SchedulerState) GetAssetsInOrg(org string) (res []*types.Asset) {
	for _, asset := range statePtr.assets[org] {
		res = append(res, asset)
	}
	return
}

// SetChecks overwrites current set of checks w/ given
func (statePtr *SchedulerState) SetChecks(checks []*types.CheckConfig) {
	statePtr.checks = make(map[string]*types.CheckConfig)
	for _, check := range checks {
		statePtr.addCheck(check)
	}
}

// SetAssets overwrites current set of assets w/ given
func (statePtr *SchedulerState) SetAssets(assets []*types.Asset) {
	statePtr.assets = make(map[string]map[string]*types.Asset)
	for _, asset := range assets {
		statePtr.addAsset(asset)
	}
}

func (statePtr *SchedulerState) addCheck(check *types.CheckConfig) {
	key := concatUniqueKey(check.Name, check.Organization, check.Environment)
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

func concatUniqueKey(args ...string) string {
	return strings.Join(args, "-")
}
