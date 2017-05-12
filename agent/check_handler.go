package agent

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/types"
)

// TODO(greg): At some point, we're going to need max parallelism.
func (a *Agent) handleCheck(payload []byte) error {
	event := &types.Event{}
	err := json.Unmarshal(payload, event)
	if err != nil {
		return err
	}

	if event.Check == nil {
		return errors.New("no check found in event")
	}

	if err := event.Check.Validate(); err != nil {
		return err
	}

	logger.Info("scheduling check execution: ", event.Check.Name)

	go a.executeCheck(event)

	return nil
}

func (a *Agent) executeCheck(event *types.Event) {
	deps := newDependencyManager(a, event.Check)

	// Inject the dependenices into PATH, LD_LIBRARY_PATH & CPATH so that they are
	// availabe when when the command is executed.
	env := os.Environ()
	env = deps.injectIntoEnv(env)

	ex := &command.Execution{
		Command: event.Check.Command,
		Env:     env,
	}
	event.Check.Executed = time.Now().Unix()

	// Ensure that all the dependencies are installed.
	deps.install()

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
