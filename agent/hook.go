package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/types"
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
				hook := a.executeHook(getHookConfig(hookName, request.Hooks))
				executedHooks = append(executedHooks, hook)
			}
		}
	}
	return executedHooks
}

func (a *Agent) executeHook(hookConfig *types.HookConfig) *types.Hook {
	if hookConfig == nil {
		return nil
	}

	// Instantiate Event and Hook
	event := &types.Event{}
	hook := &types.Hook{
		Config:   hookConfig,
		Executed: time.Now().Unix(),
	}

	// Validate that the given hook is valid.
	if err := hookConfig.Validate(); err != nil {
		a.sendFailure(event, fmt.Errorf("given hook is invalid: %s", err))
		return nil
	}

	// Instantiate the execution command
	ex := &command.Execution{
		Command: hookConfig.Command,
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
