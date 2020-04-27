package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/token"
	"github.com/sensu/sensu-go/util/environment"
	"github.com/sirupsen/logrus"
)

// ExecuteHooks executes all hooks contained in a check request based on
// the check status code of the check request
func (a *Agent) ExecuteHooks(ctx context.Context, request *corev2.CheckRequest, event *corev2.Event, assets map[string]*corev2.AssetList) []*corev2.Hook {
	executedHooks := []*corev2.Hook{}
	for _, hookList := range request.Config.CheckHooks {
		// find the hookList with the corresponding type
		if hookShouldExecute(hookList.Type, event.Check.Status) {
			// run all the hooks of that type
			for _, hookName := range hookList.Hooks {
				hookConfig := getHookConfig(hookName, request.Hooks)
				origCommand := hookConfig.Command
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
					hook := a.executeHook(ctx, hookConfig, event, assets)
					// To guard against publishing sensitive/redacted client attribute values
					// the original command value is reinstated.
					hook.Command = origCommand
					executedHooks = append(executedHooks, hook)
				}
			}
		}
	}
	return executedHooks
}

func (a *Agent) executeHook(ctx context.Context, hookConfig *corev2.HookConfig, event *corev2.Event, hookAssets map[string]*corev2.AssetList) *corev2.Hook {
	// Instantiate Hook
	hook := &corev2.Hook{
		HookConfig: *hookConfig,
		Executed:   time.Now().Unix(),
	}

	// Prepare log entry
	fields := logrus.Fields{
		"namespace": hook.Namespace,
		"hook":      hook.Name,
		"assets":    hook.RuntimeAssets,
	}

	// Match check against allow list
	var matchedEntry allowList
	var match bool
	if len(a.allowList) != 0 {
		logger.WithFields(fields).Debug("matching hook against agent allow list")
		matchedEntry, match = a.matchAllowList(hookConfig.Command)
		if !match {
			logger.WithFields(fields).Debug("hook does not match agent allow list")
			return failedHook(hook)
		}
		logger.WithFields(fields).Debug("hook matches agent allow list")
	}

	// Fetch and install all assets required for hook execution.
	logger.WithFields(fields).Debug("fetching assets for hook")
	var assetList []corev2.Asset
	if hookAssets != nil {
		if value, in := hookAssets[hook.Name]; in {
			assetList = value.Assets
		}
	}
	assets, err := asset.GetAll(ctx, a.assetGetter, assetList)
	if err != nil {
		logger.WithError(err).WithFields(fields).Error("error getting assets for hook")
		return failedHook(hook)
	}

	// Prepare environment
	env := environment.MergeEnvironments(os.Environ(), assets.Env())

	// Verify sha against the allow list
	if matchedEntry.Sha512 != "" {
		logger.WithFields(fields).Debug("matching hook sha against agent allow list")
		path, err := lookPath(strings.Split(hookConfig.Command, " ")[0], env)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("unable to find the executable path")
			return failedHook(hook)
		}
		file, err := os.Open(path)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("unable to open executable")
			return failedHook(hook)
		}
		verifier := asset.Sha512Verifier{}
		if err := verifier.Verify(file, matchedEntry.Sha512); err != nil {
			logger.WithFields(fields).WithError(err).Error("hook sha does not match agent allow list")
			return failedHook(hook)
		}
	}

	// Instantiate the execution command
	ex := command.ExecutionRequest{
		Command:      hookConfig.Command,
		Timeout:      int(hookConfig.Timeout),
		InProgress:   a.inProgress,
		InProgressMu: a.inProgressMu,
		Name:         event.Check.ObjectMeta.Name,
		Env:          env,
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

func (a *Agent) prepareHook(hookConfig *corev2.HookConfig) bool {
	if hookConfig == nil {
		return false
	}

	// Instantiate an event in case of failure
	event := &corev2.Event{
		Check: &corev2.Check{},
	}

	// Validate that the given hook is valid.
	if err := hookConfig.Validate(); err != nil {
		a.sendFailure(event, fmt.Errorf("given hook is invalid: %s", err))
		return false
	}

	if err := token.SubstituteHook(hookConfig, a.getAgentEntity()); err != nil {
		a.sendFailure(event, fmt.Errorf("error while substituting hook config tokens: %s", err))
		return false
	}

	return true
}

func getHookConfig(hookName string, hookList []corev2.HookConfig) *corev2.HookConfig {
	for _, hook := range hookList {
		if hook.Name == hookName {
			return &hook
		}
	}
	return nil
}

func hookInList(hookName string, hookList []*corev2.Hook) bool {
	for _, hook := range hookList {
		if hook.Name == hookName {
			return true
		}
	}
	return false
}

func hookShouldExecute(hookType string, status uint32) bool {
	if (hookType == strconv.FormatInt(int64(status), 10)) ||
		(hookType == "non-zero" && status != 0) ||
		(hookType == "ok" && status == 0) ||
		(hookType == "warning" && status == 1) ||
		(hookType == "critical" && status == 2) ||
		(hookType == "unknown" && (status < 0 || status > 2)) {
		return true
	}
	return false
}

func failedHook(hook *corev2.Hook) *corev2.Hook {
	hook.Status = 3
	hook.Output = "check hook command denied by the agent allow list"

	// Override the default hook status of 3 if an annotation is configured
	allowListStatus, ok := hook.Annotations[allowListOnDenyStatus]
	if ok {
		allowListValue, err := strconv.ParseInt(allowListStatus, 10, 32)
		if err == nil {
			hook.Status = int32(allowListValue)
		}
	}
	return hook
}
