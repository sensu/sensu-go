package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/dynamic"
)

// ExecuteHooks executes all hooks contained in a check request based on
// the check status code of the check request
func (a *Agent) ExecuteHooks(request *types.CheckRequest, status int) []*types.Hook {
	executedHooks := []*types.Hook{}
	for _, hookList := range request.Config.CheckHooks {
		// find the hookList with the corresponding type
		if hookShouldExecute(hookList.Type, status) {
			// run all the hooks of that type
			for _, hookName := range hookList.Hooks {
				hookConfig := getHookConfig(hookName, request.Hooks)
				if ok := a.prepareHook(hookConfig); !ok {
					// An error occured during the preparation of the hook and the error
					// has been sent back to the server. At this point we should not
					// execute the hook and wait for the next check request
					continue
				}
				// Do not duplicate hook execution for types that fall into both an exit
				// code and severity (ex. 0, ok)
				in := hookInList(hookConfig.Name, executedHooks)
				if !in {
					hook := a.executeHook(hookConfig, request.Config.Name)
					executedHooks = append(executedHooks, hook)
				}
			}
		}
	}
	return executedHooks
}

func (a *Agent) executeHook(hookConfig *types.HookConfig, check string) *types.Hook {
	// Instantiate Event and Hook
	event := &types.Event{
		Check: &types.Check{},
	}

	hook := &types.Hook{
		HookConfig: *hookConfig,
		Executed:   time.Now().Unix(),
	}

	// Instantiate the execution command
	ex := command.ExecutionRequest{
		Command:      hookConfig.Command,
		Timeout:      int(hookConfig.Timeout),
		InProgress:   a.inProgress,
		InProgressMu: a.inProgressMu,
		Name:         check,
	}

	// If stdin is true, add JSON event data to command execution.
	if hookConfig.Stdin {
		input, err := json.Marshal(event)
		if err != nil {
			a.sendFailure(event, fmt.Errorf("error marshaling json from event: %s", err))
			return nil
		}
		ex.Input = string(input)
	}

	hookExec, err := a.executor.Execute(context.Background(), ex)
	if err != nil {
		hook.Output = err.Error()
	} else {
		hook.Output = hookExec.Output
	}

	hook.Duration = hookExec.Duration
	hook.Status = int32(hookExec.Status)

	return hook
}

func (a *Agent) prepareHook(hookConfig *types.HookConfig) bool {
	if hookConfig == nil {
		return false
	}

	// Instantiate an event in case of failure
	event := &types.Event{
		Check: &types.Check{},
	}

	// Validate that the given hook is valid.
	if err := hookConfig.Validate(); err != nil {
		a.sendFailure(event, fmt.Errorf("given hook is invalid: %s", err))
		return false
	}

	// Extract the extended attributes from the entity and combine them at the
	// top-level so they can be easily accessed using token substitution
	synthesizedEntity := dynamic.Synthesize(a.getAgentEntity())

	// Substitute tokens within the check configuration with the synthesized
	// entity
	hookConfigBytes, err := TokenSubstitution(synthesizedEntity, hookConfig)
	if err != nil {
		a.sendFailure(event, err)
		return false
	}

	// Unmarshal the check configuration obtained after the token substitution
	// back into the check config struct
	if err = json.Unmarshal(hookConfigBytes, hookConfig); err != nil {
		a.sendFailure(event, fmt.Errorf("could not unmarshal the hook config: %s", err))
		return false
	}

	return true
}

func getHookConfig(hookName string, hookList []types.HookConfig) *types.HookConfig {
	for _, hook := range hookList {
		if hook.Name == hookName {
			return &hook
		}
	}
	return nil
}

func hookInList(hookName string, hookList []*types.Hook) bool {
	for _, hook := range hookList {
		if hook.Name == hookName {
			return true
		}
	}
	return false
}

func hookShouldExecute(hookType string, status int) bool {
	if (hookType == strconv.Itoa(status)) ||
		(hookType == "non-zero" && status != 0) ||
		(hookType == "ok" && status == 0) ||
		(hookType == "warning" && status == 1) ||
		(hookType == "critical" && status == 2) ||
		(hookType == "unknown" && (status < 0 || status > 2)) {
		return true
	}
	return false
}
