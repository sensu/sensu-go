package actions

import (
	"encoding/json"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	utilstrings "github.com/sensu/sensu-go/util/strings"
	"golang.org/x/net/context"
)

// checkConfigUpdateFields whitelists fields allowed to be updated for CheckConfigs
var checkConfigUpdateFields = []string{
	"Command",
	"Handlers",
	"HighFlapThreshold",
	"LowFlapThreshold",
	"Interval",
	"Publish",
	"RuntimeAssets",
	"ProxyEntityID",
	"Stdin",
	"Subscriptions",
	"CheckHooks",
	"Subdue",
	"Cron",
	"Timeout",
	"Ttl",
	"ProxyRequests",
}

var (
	adhocQueueName = "adhocRequest"
)

// CheckStore contains storage and queue info for Checks.
type CheckStore interface {
	store.CheckConfigStore
	queue.Get
}

// CheckController exposes actions in which a viewer can perform.
type CheckController struct {
	Store      CheckStore
	Policy     authorization.CheckPolicy
	checkQueue queue.Interface
}

// NewCheckController returns new CheckController
func NewCheckController(store CheckStore) CheckController {
	return CheckController{
		Store:      store,
		Policy:     authorization.Checks,
		checkQueue: store.NewQueue(adhocQueueName),
	}
}

// Query returns resources available to the viewer.
func (a CheckController) Query(ctx context.Context) ([]*types.CheckConfig, error) {
	// Fetch from store
	results, serr := a.Store.GetCheckConfigs(ctx)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	// Filter out those resources the viewer does not have access to view.
	abilities := a.Policy.WithContext(ctx)
	for i := 0; i < len(results); i++ {
		if !abilities.CanRead(results[i]) {
			results = append(results[:i], results[i+1:]...)
			i--
		}
	}

	return results, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a CheckController) Find(ctx context.Context, name string) (*types.CheckConfig, error) {
	// Fetch from store
	result, serr := a.Store.GetCheckConfigByName(ctx, name)

	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	// Verify user has permission to view
	abilities := a.Policy.WithContext(ctx)
	if result != nil && abilities.CanRead(result) {
		return result, nil
	}

	return nil, NewErrorf(NotFound)
}

// Create instatiates, validates and persists new resource if viewer has access.
func (a CheckController) Create(ctx context.Context, newCheck types.CheckConfig) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &newCheck)
	abilities := a.Policy.WithContext(ctx)

	// Check for existing
	if e, err := a.Store.GetCheckConfigByName(ctx, newCheck.Name); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Verify viewer can make change
	if yes := abilities.CanCreate(&newCheck); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Validate
	if err := newCheck.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.Store.UpdateCheckConfig(ctx, &newCheck); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Update validates and persists changes to a resource if viewer has access.
func (a CheckController) Update(ctx context.Context, given types.CheckConfig) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &given)
	abilities := a.Policy.WithContext(ctx)

	// Find existing check
	check, err := a.Store.GetCheckConfigByName(ctx, given.Name)
	if err != nil {
		return NewError(InternalErr, err)
	} else if check == nil {
		return NewErrorf(NotFound)
	}

	// Verify viewer can make change
	if yes := abilities.CanUpdate(check); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Copy
	copyFields(check, &given, checkConfigUpdateFields...)

	// Validate
	if err := check.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist Changes
	if serr := a.Store.UpdateCheckConfig(ctx, check); serr != nil {
		return NewError(InternalErr, serr)
	}

	return nil
}

// Destroy removes a resource if viewer has access.
func (a CheckController) Destroy(ctx context.Context, name string) error {
	abilities := a.Policy.WithContext(ctx)

	// Verify user has permission
	if yes := abilities.CanDelete(); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Fetch from store
	result, serr := a.Store.GetCheckConfigByName(ctx, name)
	if serr != nil {
		return NewError(InternalErr, serr)
	} else if result == nil {
		return NewErrorf(NotFound)
	}

	// Remove from store
	if err := a.Store.DeleteCheckConfigByName(ctx, result.Name); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// AddCheckHook adds an association between a hook and a check
func (a CheckController) AddCheckHook(ctx context.Context, check string, checkHook types.HookList) error {
	return a.findAndUpdateCheckConfig(ctx, check, func(check *types.CheckConfig) error {
		var exists bool
		for i, r := range check.CheckHooks {
			if r.Type == checkHook.Type {
				exists = true
				hookList := check.CheckHooks[i].Hooks
				// if the type already exists in the check's check hooks, only append the hook names provided
				for _, h := range checkHook.Hooks {
					if !utilstrings.InArray(h, hookList) {
						// only add hook names that don't already exist in list
						hookList = append(hookList, h)
					}
				}
				check.CheckHooks[i].Hooks = hookList
				break
			}
		}

		if !exists {
			// if the type doesn't alrady exist, just add the bulk check hook
			check.CheckHooks = append(check.CheckHooks, checkHook)
		}
		return nil
	})
}

// RemoveCheckHook removes an association between a hook and a check
func (a CheckController) RemoveCheckHook(ctx context.Context, checkName string, hookType string, hookName string) error {
	return a.findAndUpdateCheckConfig(ctx, checkName, func(check *types.CheckConfig) error {
		for i, r := range check.CheckHooks {
			if r.Type == hookType {
				hookList := check.CheckHooks[i].Hooks
				for j, h := range hookList {
					if h == hookName {
						check.CheckHooks[i].Hooks = append(hookList[:j], hookList[j+1:]...)
						if len(check.CheckHooks[i].Hooks) == 0 {
							// if the type contains no hook names, remove type
							check.CheckHooks = append(check.CheckHooks[:i], check.CheckHooks[i+1:]...)
						}
						return nil
					}
				}
			}
		}

		return NewErrorf(NotFound)
	})
}

func (a CheckController) findCheckConfig(ctx context.Context, name string) (*types.CheckConfig, error) {
	result, serr := a.Store.GetCheckConfigByName(ctx, name)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	} else if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

func (a CheckController) updateCheckConfig(ctx context.Context, check *types.CheckConfig) error {
	if err := a.Store.UpdateCheckConfig(ctx, check); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

func (a CheckController) findAndUpdateCheckConfig(
	ctx context.Context,
	name string,
	configureFn func(*types.CheckConfig) error,
) error {
	// Find
	check, serr := a.findCheckConfig(ctx, name)
	if serr != nil {
		return serr
	}

	// Verify viewer can make change
	abilities := a.Policy.WithContext(ctx)
	if yes := abilities.CanUpdate(check); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Configure
	if err := configureFn(check); err != nil {
		return err
	}

	// Validate
	if err := check.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Update
	return a.updateCheckConfig(ctx, check)
}

// QueueAdhocRequest takes a check request and adds it to the queue for
// processing.
func (a CheckController) QueueAdhocRequest(ctx context.Context, name string, adhocRequest *types.AdhocRequest) error {
	checkConfig, err := a.Find(ctx, name)
	if err != nil {
		return err
	}

	// Adjust context
	ctx = addOrgEnvToContext(ctx, checkConfig)
	abilities := a.Policy.WithContext(ctx)

	// Verify viewer can make change
	if yes := abilities.CanCreate(checkConfig); !yes {
		return NewErrorf(PermissionDenied)
	}

	// if there are subscriptions, update the check with the provided subscriptions;
	// otherwise, use what the check already has
	if len(adhocRequest.Subscriptions) > 0 {
		checkConfig.Subscriptions = adhocRequest.Subscriptions
	}

	// finally, add the check to the queue
	marshaledCheck, err := json.Marshal(checkConfig)
	if err != nil {
		return err
	}
	err = a.checkQueue.Enqueue(ctx, string(marshaledCheck))
	return err
}
