package schedulerd

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// StateManagerStore specifies the storage requirements for StateManagers.
type StateManagerStore interface {
	store.AssetStore
	store.CheckConfigStore
	store.EntityStore
	store.HookConfigStore
}

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
func NewStateManager(store StateManagerStore) *StateManager {
	manager := &StateManager{
		OnChecksChange: func(state *SchedulerState) {},

		state: &atomic.Value{},
		mutex: &sync.Mutex{},
	}
	manager.Swap(&SchedulerState{})

	manager.synchronizer = NewSynchronizeStateScheduler(
		SynchronizeMinInterval,
		&SynchronizeChecks{
			Store:    store,
			OnUpdate: manager.updateChecks,
		},
		&SynchronizeAssets{
			Store:    store,
			OnUpdate: manager.updateAssets,
		},
		&SynchronizeHooks{
			Store:    store,
			OnUpdate: manager.updateHooks,
		},
		&SynchronizeEntities{
			Store:    store,
			OnUpdate: manager.updateEntities,
		},
	)

	return manager
}

// Start keeping state synchronized
func (mngrPtr *StateManager) Start(ctx context.Context) {
	mngrPtr.synchronizer.Start(ctx)
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

func (mngrPtr *StateManager) updateHooks(hooks []*types.HookConfig) {
	mngrPtr.updateState(func(state *SchedulerState) {
		state.SetHooks(hooks)
	})
}

func (mngrPtr *StateManager) updateEntities(entities []*types.Entity) {
	mngrPtr.updateState(func(state *SchedulerState) {
		state.SetEntities(entities)
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
	checks   map[string]*types.CheckConfig
	assets   map[string]map[string]*types.Asset
	hooks    map[string]map[string]*types.HookConfig
	entities map[string]map[string]*types.Entity
}

// GetCheck returns check given name and organization
func (statePtr *SchedulerState) GetCheck(name, org, env string) *types.CheckConfig {
	key := concatUniqueKey(name, org, env)
	return statePtr.checks[key]
}

// GetAssetsInNamespace returns all assets associated given organization
func (statePtr *SchedulerState) GetAssetsInNamespace(org string) (res []*types.Asset) {
	for _, asset := range statePtr.assets[org] {
		res = append(res, asset)
	}
	return
}

// GetHooksInNamespace returns all hooks associated given organization
// and environment
func (statePtr *SchedulerState) GetHooksInNamespace(org string, env string) (res []*types.HookConfig) {
	orgEnv := concatUniqueKey(org, env)
	for _, hook := range statePtr.hooks[orgEnv] {
		res = append(res, hook)
	}
	return
}

// GetEntitiesInNamespace returns all entities associated given organization
// and environment
func (statePtr *SchedulerState) GetEntitiesInNamespace(org string, env string) (res []*types.Entity) {
	orgEnv := concatUniqueKey(org, env)
	for _, entity := range statePtr.entities[orgEnv] {
		res = append(res, entity)
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

// SetHooks overwrites current set of hooks w/ given
func (statePtr *SchedulerState) SetHooks(hooks []*types.HookConfig) {
	statePtr.hooks = make(map[string]map[string]*types.HookConfig)
	for _, hook := range hooks {
		statePtr.addHook(hook)
	}
}

// SetEntities overwrites current set of entities w/ given
func (statePtr *SchedulerState) SetEntities(entities []*types.Entity) {
	statePtr.entities = make(map[string]map[string]*types.Entity)
	for _, entity := range entities {
		statePtr.addEntity(entity)
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

func (statePtr *SchedulerState) addHook(hook *types.HookConfig) {
	org := hook.Organization
	env := hook.Environment
	orgEnv := concatUniqueKey(org, env)
	if hookMap := statePtr.hooks[orgEnv]; hookMap == nil {
		statePtr.hooks[orgEnv] = make(map[string]*types.HookConfig)
	}

	key := concatUniqueKey(hook.Name, org, env)
	statePtr.hooks[orgEnv][key] = hook
}

func (statePtr *SchedulerState) addEntity(entity *types.Entity) {
	org := entity.Organization
	env := entity.Environment
	orgEnv := concatUniqueKey(org, env)
	if entityMap := statePtr.entities[orgEnv]; entityMap == nil {
		statePtr.entities[orgEnv] = make(map[string]*types.Entity)
	}

	key := concatUniqueKey(entity.ID, org, env)
	statePtr.entities[orgEnv][key] = entity
}

func concatUniqueKey(args ...string) string {
	return strings.Join(args, "-")
}
