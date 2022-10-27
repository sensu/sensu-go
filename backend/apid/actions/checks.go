package actions

import (
	"encoding/json"

	"context"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/types"
	utilstrings "github.com/sensu/sensu-go/util/strings"
)

var (
	adhocQueueName = "adhocRequest"
)

// CheckController exposes actions which a viewer can perform.
type CheckController struct {
	store      storev2.Interface
	checkQueue types.Queue
}

// NewCheckController returns new CheckController
func NewCheckController(store storev2.Interface, getter types.QueueGetter) CheckController {
	return CheckController{
		store:      store,
		checkQueue: getter.GetQueue(adhocQueueName),
	}
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a CheckController) Find(ctx context.Context, name string) (*corev2.CheckConfig, error) {
	// Fetch from store
	var check corev2.CheckConfig
	check.Name = name
	check.Namespace = corev2.ContextNamespace(ctx)
	req := storev2.NewResourceRequestFromResource(&check)
	wrapper, err := a.store.Get(ctx, req)
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			return nil, NewErrorf(NotFound)
		} else {
			return nil, NewError(InternalErr, err)
		}
	}

	if err := wrapper.UnwrapInto(&check); err != nil {
		return nil, NewError(InternalErr, err)
	}

	return &check, nil
}

// AddCheckHook adds an association between a hook and a check
func (a CheckController) AddCheckHook(ctx context.Context, check string, checkHook corev2.HookList) error {
	return a.findAndUpdateCheckConfig(ctx, check, func(check *corev2.CheckConfig) error {
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
	return a.findAndUpdateCheckConfig(ctx, checkName, func(check *corev2.CheckConfig) error {
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

func (a CheckController) updateCheckConfig(ctx context.Context, check *corev2.CheckConfig) error {
	req := storev2.NewResourceRequestFromResource(check)
	wrapper, err := storev2.WrapResource(check)
	if err != nil {
		return NewError(InternalErr, err)
	}
	if err := a.store.CreateOrUpdate(ctx, req, wrapper); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

func (a CheckController) findAndUpdateCheckConfig(
	ctx context.Context,
	name string,
	configureFn func(*corev2.CheckConfig) error,
) error {
	// Find
	check, err := a.Find(ctx, name)
	if err != nil {
		return err
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
func (a CheckController) QueueAdhocRequest(ctx context.Context, name string, adhocRequest *corev2.AdhocRequest) error {
	checkConfig, err := a.Find(ctx, name)
	if err != nil {
		return err
	}

	// Adjust context
	ctx = corev2.SetContextFromResource(ctx, checkConfig)

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
