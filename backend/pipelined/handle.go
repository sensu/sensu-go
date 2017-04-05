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
	for _, handlerName := range event.Check.Handlers {
		handler, err := p.Store.GetHandlerByName(handlerName)

		if err != nil {
			log.Println("pipelined failed to retrieve a handler: ", err.Error())
			continue
		}

		eventData, err := p.mutateEvent(handler, event)

		if err != nil {
			continue
		}

		switch handler.Type {
		case "pipe":
			p.pipeHandler(handler, eventData)
		default:
			return errors.New("unknown handler type")
		}
	}

	return nil
}

func (p *Pipelined) pipeHandler(handler *types.Handler, eventData []byte) (*command.Execution, error) {
	handlerExec := &command.Execution{}

	handlerExec.Command = handler.Pipe.Command
	handlerExec.Timeout = handler.Pipe.Timeout

	handlerExec.Input = string(eventData[:])

	result, err := command.ExecuteCommand(context.Background(), handlerExec)

	if err != nil {
		log.Println("pipelined failed to execute event pipe handler: ", err.Error())
	} else {
		log.Printf("pipelined executed event pipe handler: status: %x output: %s", result.Status, result.Output)
	}

	return result, err
}
