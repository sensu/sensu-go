package agent

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/types"
)

// TODO(greg): At some point, we're going to need max parallelism.
func (a *Agent) handleCheck(payload []byte) error {
	request := &types.CheckRequest{}
	err := json.Unmarshal(payload, request)
	if err != nil {
		return err
	}

	if request == nil {
		return errors.New("given check configuration appears invalid")
	}

	checkConfig := request.Config
	if err := checkConfig.Validate(); err != nil {
		return err
	}

	logger.Info("scheduling check execution: ", checkConfig.Name)

	go a.executeCheck(request)

	return nil
}

func (a *Agent) executeCheck(request *types.CheckRequest) {
	checkConfig := request.Config
	checkAssets := request.ExpandedAssets

	// Ensure that the asset manager is aware of all the assets required to
	// execute the given check.
	assets := a.assetManager.RegisterSet(checkConfig.RuntimeAssets)

	ex := &command.Execution{
		// Inject the dependenices into PATH, LD_LIBRARY_PATH & CPATH so that they are
		// availabe when when the command is executed.
		Env:     assets.Env(),
		Command: checkConfig.Command,
	}
	event := &types.Event{
		Check: &types.Check{
			Config:   checkConfig,
			Executed: time.Now().Unix(),
		},
	}

	// Ensure that all the dependencies are installed.
	if err := assets.InstallAll(); err != nil {
		logger.Error("error installing dependencies: ", err.Error())
		return
	}

	_, err := command.ExecuteCommand(context.Background(), ex)
	if err != nil {
		event.Check.Output = err.Error()
	} else {
		event.Check.Output = ex.Output
	}

	event.Check.Duration = ex.Duration
	event.Check.Status = ex.Status

	event.Entity = a.getAgentEntity()
	event.Timestamp = time.Now().Unix()

	msg, err := json.Marshal(event)
	if err != nil {
		logger.Error("error marshaling check result: ", err.Error())
		return
	}

	a.sendMessage(types.EventType, msg)
}
