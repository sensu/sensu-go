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
)

// TODO(greg): At some point, we're going to need max parallelism.
func (a *Agent) handleCheck(payload []byte) error {
	request := &types.CheckRequest{}
	if err := json.Unmarshal(payload, request); err != nil {
		return err
	} else if request == nil {
		return errors.New("given check configuration appears invalid")
	}

	logger.Info("scheduling check execution: ", request.Config.Name)
	go a.executeCheck(request)

	return nil
}

func (a *Agent) executeCheck(request *types.CheckRequest) {
	checkConfig := request.Config
	checkAssets := request.Assets

	// Instantiate Event
	event := &types.Event{
		Check: &types.Check{
			Config:   checkConfig,
			Executed: time.Now().Unix(),
		},
	}

	// Validate that the given check is valid.
	if err := checkConfig.Validate(); err != nil {
		a.sendFailure(event, fmt.Errorf("given check is invalid: %s", err))
		return
	}

	// Ensure that the asset manager is aware of all the assets required to
	// execute the given check.
	assets := a.assetManager.RegisterSet(checkAssets)

	// Inject the dependenices into PATH, LD_LIBRARY_PATH & CPATH so that they are
	// availabe when when the command is executed.
	ex := &command.Execution{
		Env:     assets.Env(),
		Command: checkConfig.Command,
	}

	// Ensure that all the dependencies are installed.
	if err := assets.InstallAll(); err != nil {
		a.sendFailure(event, fmt.Errorf("error install dependencies: %s", err))
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

	msg, err := json.Marshal(event)
	if err != nil {
		logger.Error("error marshaling check result: ", err.Error())
		return
	}

	a.sendMessage(transport.MessageTypeEvent, msg)
}

func (a *Agent) sendFailure(event *types.Event, err error) {
	event.Check.Output = err.Error()
	event.Check.Status = 3
	event.Timestamp = time.Now().Unix()

	if msg, err := json.Marshal(event); err != nil {
		logger.Error("error marshaling check failure: ", err.Error())
	} else {
		a.sendMessage(transport.MessageTypeEvent, msg)
	}
}
