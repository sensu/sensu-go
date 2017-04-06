// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"errors"
	"log"

	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/types"
)

// handleEvent takes a Sensu event through a Sensu pipeline, filters
// -> mutator -> handler. An event may have one or more handlers. Most
// errors are only logged and used for flow control, they will not
// interupt event handling.
func (p *Pipelined) handleEvent(event *types.Event) error {
	handlers, err := p.expandHandlers(event.Check.Handlers, 1)

	if err != nil {
		return err
	}

	for _, handler := range handlers {
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

// expandHandlers turns a list of Sensu handler names into a list of
// handlers, while expanding handler sets with support for some
// nesting. Handlers are fetched from etcd.
func (p *Pipelined) expandHandlers(handlers []string, level int) (map[string]*types.Handler, error) {
	if level > 3 {
		return nil, errors.New("handler sets cannot be deeply nested")
	}

	expanded := map[string]*types.Handler{}

	for _, handlerName := range handlers {
		handler, err := p.Store.GetHandlerByName(handlerName)

		if handler == nil {
			if err != nil {
				log.Println("pipelined failed to retrieve a handler: ", err.Error())
			} else {
				log.Println("pipelined failed to retrieve a handler: ", handlerName)
			}
			continue
		}

		if handler.Type == "set" {
			level++
			setHandlers, err := p.expandHandlers(handler.Handlers, level)

			if err != nil {
				log.Println("pipelined failed to expand handler set: ", err.Error())
			} else {
				for name, setHandler := range setHandlers {
					if _, ok := expanded[name]; !ok {
						expanded[name] = setHandler
					}
				}
			}
		} else {
			if _, ok := expanded[handler.Name]; !ok {
				expanded[handler.Name] = handler
			}
		}
	}

	return expanded, nil
}

// pipeHandler fork/executes a child process for a Sensu pipe handler
// command and writes the mutated eventData to it via STDIN.
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
