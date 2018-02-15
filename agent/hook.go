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
				hook := a.executeHook(hookConfig)
				executedHooks = append(executedHooks, hook)
			}
		}
	}
	return executedHooks
}

func (a *Agent) executeHook(hookConfig *types.HookConfig) *types.Hook {
	// Instantiate Event and Hook
	event := &types.Event{
		Check: &types.Check{},
	}

	hook := &types.Hook{
		HookConfig: *hookConfig,
		Executed:   time.Now().Unix(),
	}

	// Instantiate the execution command
	ex := &command.Execution{
		Command: hookConfig.Command,
		Timeout: int(hookConfig.Timeout),
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

	if _, err := command.ExecuteCommand(context.Background(), ex); err != nil {
		hook.Output = err.Error()
	} else {
		hook.Output = ex.Output
	}

	hook.Duration = ex.Duration
	hook.Status = int32(ex.Status)

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
	synthesizedEntity, err := dynamic.Synthesize(a.getAgentEntity())
	if err != nil {
		a.sendFailure(event, fmt.Errorf("could not synthesize the entity: %s", err))
		return false
	}

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
