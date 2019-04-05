package actions

import (
	"encoding/json"

	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	utilstrings "github.com/sensu/sensu-go/util/strings"
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
	"ProxyEntityName",
	"Stdin",
	"Subscriptions",
	"CheckHooks",
	"Subdue",
	"Cron",
	"Timeout",
	"Ttl",
	"ProxyRequests",
	"OutputMetricFormat",
	"OutputMetricHandlers",
}

var (
	adhocQueueName = "adhocRequest"
)

// CheckController exposes actions which a viewer can perform.
type CheckController struct {
	store      store.CheckConfigStore
	checkQueue types.Queue
}

// NewCheckController returns new CheckController
func NewCheckController(store store.CheckConfigStore, getter types.QueueGetter) CheckController {
	return CheckController{
		store:      store,
		checkQueue: getter.GetQueue(adhocQueueName),
	}
}

// List returns resources available to the viewer.
func (a CheckController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	// Fetch from store
	results, err := a.store.GetCheckConfigs(ctx, pred)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	resources := make([]corev2.Resource, len(results))
	for i, v := range results {
		resources[i] = corev2.Resource(v)
	}

	return resources, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a CheckController) Find(ctx context.Context, name string) (*types.CheckConfig, error) {
	// Fetch from store
	result, serr := a.store.GetCheckConfigByName(ctx, name)

	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}
	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

// Create instantiates, validates and persists new resource if viewer has access.
func (a CheckController) Create(ctx context.Context, newCheck types.CheckConfig) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &newCheck)

	// Check for existing
	if e, err := a.store.GetCheckConfigByName(ctx, newCheck.Name); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Validate
	if err := newCheck.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.store.UpdateCheckConfig(ctx, &newCheck); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// CreateOrReplace instatiates and persists new resource if viewer has access.
func (a CheckController) CreateOrReplace(ctx context.Context, newCheck types.CheckConfig) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &newCheck)

	// Validate
	if err := newCheck.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.store.UpdateCheckConfig(ctx, &newCheck); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Destroy removes a resource if viewer has access.
func (a CheckController) Destroy(ctx context.Context, name string) error {
	// Fetch from store
	result, serr := a.store.GetCheckConfigByName(ctx, name)
	if serr != nil {
		return NewError(InternalErr, serr)
	} else if result == nil {
		return NewErrorf(NotFound)
	}

	// Remove from store
	if err := a.store.DeleteCheckConfigByName(ctx, result.Name); err != nil {
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
	result, serr := a.store.GetCheckConfigByName(ctx, name)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	} else if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

func (a CheckController) updateCheckConfig(ctx context.Context, check *types.CheckConfig) error {
	if err := a.store.UpdateCheckConfig(ctx, check); err != nil {
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
