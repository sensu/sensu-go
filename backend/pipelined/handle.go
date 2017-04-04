// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"errors"
	"log"

	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/types"
)

func (p *Pipelined) handleEvent(event *types.Event) error {
	handlerPipe := &types.HandlerPipe{
		Command: "cat",
	}

	handler := &types.Handler{
		Type: "pipe",
		Pipe: *handlerPipe,
	}

	eventData, err := p.mutateEvent(handler, event)

	if err != nil {
		return err
	}

	if handler.Type == "pipe" {
		handlerExec, err := p.pipeHandler(handler, eventData)

		if err != nil {
			return err
		}

		log.Printf("executed event pipe handler: status: %x output: %s", handlerExec.Status, handlerExec.Output)
	} else {
		// We MUST validate handler type.
		return errors.New("unknown handler type")
	}

	return nil
}

func (p *Pipelined) pipeHandler(handler *types.Handler, eventData []byte) (*command.Execution, error) {
	handlerExec := &command.Execution{}

	handlerExec.Command = handler.Pipe.Command
	handlerExec.Timeout = handler.Pipe.Timeout

	handlerExec.Input = string(eventData[:])

	result, err := command.ExecuteCommand(context.Background(), handlerExec)

	return result, err
}
