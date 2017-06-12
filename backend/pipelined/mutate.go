// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/types"
)

// mutateEvent mutates (transforms) a Sensu event into a serialized
// format (byte slice) to be provided to a Sensu event handler.
func (p *Pipelined) mutateEvent(handler *types.Handler, event *types.Event) ([]byte, error) {
	if handler.Mutator == "" {
		eventData, err := p.jsonMutator(event)

		if err != nil {
			logger.Error("pipelined failed to mutate an event: ", err.Error())
			return nil, err
		}

		return eventData, nil
	}

	mutator, err := p.Store.GetMutatorByName(event.Entity.Organization, handler.Mutator)

	if mutator == nil {
		if err != nil {
			logger.Error("pipelined failed to retrieve a mutator: ", err.Error())
		} else {
			logger.Error("pipelined failed to retrieve a mutator: name= ", handler.Mutator)
		}
		return nil, err
	}

	eventData, err := p.pipeMutator(mutator, event)

	if err != nil {
		logger.Error("pipelined failed to mutate an event: ", err.Error())
		return nil, err
	}

	return eventData, nil
}

// jsonMutator produces the JSON encoding of the Sensu event. This
// mutator is used when a Sensu handler does not specify one.
func (p *Pipelined) jsonMutator(event *types.Event) ([]byte, error) {
	eventData, err := json.Marshal(event)

	if err != nil {
		return nil, err
	}

	return eventData, nil
}

// pipeMutator fork/executes a child process for a Sensu mutator
// command, writes the JSON encoding of the Sensu event to it via
// STDIN, and captures the command output (STDOUT/ERR) to be used as
// the mutated event data for a Sensu event handler.
func (p *Pipelined) pipeMutator(mutator *types.Mutator, event *types.Event) ([]byte, error) {
	mutatorExec := &command.Execution{}

	mutatorExec.Command = mutator.Command
	mutatorExec.Timeout = mutator.Timeout
	mutatorExec.Env = mutator.Env

	eventData, err := json.Marshal(event)

	if err != nil {
		return nil, err
	}

	mutatorExec.Input = string(eventData[:])

	result, err := command.ExecuteCommand(context.Background(), mutatorExec)

	if err != nil {
		return nil, err
	} else if result.Status != 0 {
		return nil, errors.New("pipe mutator execution returned non-zero exit status")
	}

	logger.Debug("pipelined executed event pipe mutator: status=%x output=%s", result.Status, result.Output)

	return []byte(result.Output), nil
}
