// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

// mutateEvent mutates (transforms) a Sensu event into a serialized
// format (byte slice) to be provided to a Sensu event handler.
func (p *Pipelined) mutateEvent(handler *types.Handler, event *types.Event) ([]byte, error) {
	if handler.Mutator == "" {
		eventData, err := p.jsonMutator(event)

		if err != nil {
			logger.WithError(err).Error("pipelined failed to mutate an event")
			return nil, err
		}

		return eventData, nil
	}

	if handler.Mutator == "only_check_output" {
		eventData := p.onlyCheckOutputMutator(event)

		return eventData, nil
	}

	ctx := context.WithValue(context.Background(), types.OrganizationKey, event.Entity.Organization)
	ctx = context.WithValue(ctx, types.EnvironmentKey, event.Entity.Environment)
	mutator, err := p.store.GetMutatorByName(ctx, handler.Mutator)
	if err != nil {
		logger.WithError(err).Error("pipelined failed to retrieve a mutator")
		return nil, err
	}

	if mutator == nil {
		// Check to see if there is an extension matching the mutator
		extension, err := p.store.GetExtension(ctx, handler.Mutator)
		if err != nil {
			if err == store.ErrNoExtension {
				return nil, nil
			}
			return nil, err
		}
		executor, err := p.extensionExecutor(extension)
		if err != nil {
			return nil, err
		}
		eventData, err := executor.MutateEvent(event)
		if err != nil {
			return nil, err
		}
		return eventData, nil
	}

	eventData, err := p.pipeMutator(mutator, event)

	if err != nil {
		logger.WithError(err).Error("pipelined failed to mutate an event")
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

// onlyCheckOutputMutator returns only the check output from the Sensu
// event. This mutator is considered to be "built-in" (1.x parity), it
// is most commonly used by tcp/udp handlers (e.g. influxdb). This
// mutator can probably be removed/replaced when 2.0 has extension
// support.
func (p *Pipelined) onlyCheckOutputMutator(event *types.Event) []byte {
	return []byte(event.Check.Output)
}

// pipeMutator fork/executes a child process for a Sensu mutator
// command, writes the JSON encoding of the Sensu event to it via
// STDIN, and captures the command output (STDOUT/ERR) to be used as
// the mutated event data for a Sensu event handler.
func (p *Pipelined) pipeMutator(mutator *types.Mutator, event *types.Event) ([]byte, error) {
	mutatorExec := &command.Execution{}

	mutatorExec.Command = mutator.Command
	mutatorExec.Timeout = int(mutator.Timeout)
	mutatorExec.Env = mutator.EnvVars

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

	logger.WithFields(logrus.Fields{
		"status": result.Status,
		"output": result.Output,
	}).Debug("pipelined executed event pipe mutator")

	return []byte(result.Output), nil
}
