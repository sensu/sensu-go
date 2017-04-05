// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/types"
)

func (p *Pipelined) mutateEvent(handler *types.Handler, event *types.Event) ([]byte, error) {
	if handler.Mutator == "" {
		eventData, err := p.jsonMutator(event)

		if err != nil {
			log.Println("pipelined failed to mutate an event: ", err.Error())
			return nil, err
		}

		return eventData, nil
	}

	mutator, err := p.Store.GetMutatorByName(handler.Mutator)

	if err != nil {
		log.Println("pipelined failed to retrieve a mutator: ", err.Error())
		return nil, err
	}

	eventData, err := p.pipeMutator(mutator, event)

	if err != nil {
		log.Println("pipelined failed to mutate an event: ", err.Error())
		return nil, err
	}

	return eventData, nil
}

func (p *Pipelined) jsonMutator(event *types.Event) ([]byte, error) {
	eventData, err := json.Marshal(event)

	if err != nil {
		return nil, err
	}

	return eventData, nil
}

func (p *Pipelined) pipeMutator(mutator *types.Mutator, event *types.Event) ([]byte, error) {
	mutatorExec := &command.Execution{}

	mutatorExec.Command = mutator.Command
	mutatorExec.Timeout = mutator.Timeout

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
	} else {
		log.Printf("pipelined executed event pipe mutator: status: %x output: %s", result.Status, result.Output)
	}

	return []byte(result.Output), nil
}
