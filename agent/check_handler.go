package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/dynamic"
	utilstrings "github.com/sensu/sensu-go/util/strings"
)

// tracks the check names that are still executing
var inProgress []string

// TODO(greg): At some point, we're going to need max parallelism.
func (a *Agent) handleCheck(payload []byte) error {
	request := &types.CheckRequest{}
	if err := json.Unmarshal(payload, request); err != nil {
		return err
	} else if request == nil {
		return errors.New("given check configuration appears invalid")
	}

	// only schedule check execution if its not already in progress
	// ** check hooks are part of a checks execution
	if !utilstrings.InArray(request.Config.Name, inProgress) {
		logger.Info("scheduling check execution: ", request.Config.Name)

		if ok := a.prepareCheck(request.Config); !ok {
			// An error occured during the preparation of the check and the error has
			// been sent back to the server. At this point we should not execute the
			// check and wait for the next check request
			return nil
		}

		go a.executeCheck(request)
	}

	return nil
}

func (a *Agent) executeCheck(request *types.CheckRequest) {
	inProgress = append(inProgress, request.Config.Name)
	defer func() {
		inProgress = utilstrings.Remove(request.Config.Name, inProgress)
	}()

	checkConfig := request.Config
	checkAssets := request.Assets
	checkHooks := request.Hooks

	// Instantiate Event
	event := &types.Event{
		Check: &types.Check{
			Config:   checkConfig,
			Executed: time.Now().Unix(),
		},
	}

	// Ensure that the asset manager is aware of all the assets required to
	// execute the given check.
	assets := a.assetManager.RegisterSet(checkAssets)

	// Inject the dependenices into PATH, LD_LIBRARY_PATH & CPATH so that they are
	// availabe when when the command is executed.
	ex := &command.Execution{
		Env:     assets.Env(),
		Command: checkConfig.Command,
		Timeout: int(checkConfig.Timeout),
	}

	// If stdin is true, add JSON event data to command execution.
	if checkConfig.Stdin {
		input, err := json.Marshal(event)
		if err != nil {
			a.sendFailure(event, fmt.Errorf("error marshaling json from event: %s", err))
			return
		}
		ex.Input = string(input)
	}

	// Ensure that all the dependencies are installed.
	if err := assets.InstallAll(); err != nil {
		a.sendFailure(event, fmt.Errorf("error installing dependencies: %s", err))
		return
	}

	if _, err := command.ExecuteCommand(context.Background(), ex); err != nil {
		event.Check.Output = err.Error()
	} else {
		event.Check.Output = ex.Output
	}

	event.Check.Duration = ex.Duration
	event.Check.Status = int32(ex.Status)

	event.Entity = a.getAgentEntity()
	event.Timestamp = time.Now().Unix()

	if len(checkHooks) != 0 {
		event.Hooks = a.ExecuteHooks(request, ex.Status)
	}

	msg, err := json.Marshal(event)
	if err != nil {
		logger.Error("error marshaling check result: ", err.Error())
		return
	}

	a.sendMessage(transport.MessageTypeEvent, msg)
}

// prepareCheck prepares a check before its execution by validating the
// configuration and performing token substitution. A boolean value is returned,
// indicathing whether the check should be executed or not
func (a *Agent) prepareCheck(check *types.CheckConfig) bool {
	// Instantiate an event in case of failure
	event := &types.Event{
		Check: &types.Check{
			Config:   check,
			Executed: time.Now().Unix(),
		},
	}

	// Validate that the given check is valid.
	if err := check.Validate(); err != nil {
		a.sendFailure(event, fmt.Errorf("given check is invalid: %s", err))
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
	checkBytes, err := tokenSubstitution(synthesizedEntity, check)
	if err != nil {
		a.sendFailure(event, err)
		return false
	}

	// Unmarshal the check configuration obtained after the token substitution
	// back into the check config struct
	err = json.Unmarshal(checkBytes, check)
	if err != nil {
		a.sendFailure(event, fmt.Errorf("could not unmarshal the check: %s", err))
		return false
	}

	return true
}

func (a *Agent) sendFailure(event *types.Event, err error) {
	event.Check.Output = err.Error()
	event.Check.Status = 3
	event.Entity = a.getAgentEntity()
	event.Timestamp = time.Now().Unix()

	if msg, err := json.Marshal(event); err != nil {
		logger.Error("error marshaling check failure: ", err.Error())
	} else {
		a.sendMessage(transport.MessageTypeEvent, msg)
	}
}
