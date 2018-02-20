// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/types"
)

const (
	// DefaultSocketTimeout specifies the default socket dial
	// timeout in seconds for TCP and UDP handlers.
	DefaultSocketTimeout uint32 = 60
)

// handleEvent takes a Sensu event through a Sensu pipeline, filters
// -> mutator -> handler. An event may have one or more handlers. Most
// errors are only logged and used for flow control, they will not
// interupt event handling.
func (p *Pipelined) handleEvent(event *types.Event) error {
	ctx := context.WithValue(context.Background(), types.OrganizationKey, event.Entity.Organization)
	ctx = context.WithValue(ctx, types.EnvironmentKey, event.Entity.Environment)

	var handlerList []string

	if event.HasCheck() {
		handlerList = append(handlerList, event.Check.Handlers...)
	}

	if event.HasMetrics() {
		handlerList = append(handlerList, event.Metrics.Handlers...)
	}

	handlers, err := p.expandHandlers(ctx, handlerList, 1)

	if err != nil {
		return err
	}

	for _, handler := range handlers {
		filtered := p.filterEvent(handler, event)

		if filtered {
			logger.WithFields(logrus.Fields{
				"check":        event.Check.Name,
				"entity":       event.Entity.ID,
				"organization": event.Entity.Organization,
				"environment":  event.Entity.Environment,
			}).Debug("event filtered")
			continue
		}

		eventData, err := p.mutateEvent(handler, event)

		if err != nil {
			continue
		}

		logger.Debugf("sending event: %s to handler: %s", eventData, handler.Name)

		switch handler.Type {
		case "pipe":
			if _, err := p.pipeHandler(handler, eventData); err != nil {
				logger.Error(err)
			}
		case "tcp", "udp":
			if _, err := p.socketHandler(handler, eventData); err != nil {
				logger.Error(err)
			}
		default:
			return errors.New("unknown handler type")
		}
	}

	return nil
}

// expandHandlers turns a list of Sensu handler names into a list of
// handlers, while expanding handler sets with support for some
// nesting. Handlers are fetched from etcd.
func (p *Pipelined) expandHandlers(ctx context.Context, handlers []string, level int) (map[string]*types.Handler, error) {
	if level > 3 {
		return nil, errors.New("handler sets cannot be deeply nested")
	}

	expanded := map[string]*types.Handler{}

	for _, handlerName := range handlers {
		handler, err := p.Store.GetHandlerByName(ctx, handlerName)

		if handler == nil {
			if err != nil {
				logger.Error("pipelined failed to retrieve a handler: ", err.Error())
			} else {
				logger.Error("pipelined failed to retrieve a handler: name= ", handlerName)
			}
			continue
		}

		if handler.Type == "set" {
			level++
			setHandlers, err := p.expandHandlers(ctx, handler.Handlers, level)

			if err != nil {
				logger.Error("pipelined failed to expand handler set: ", err.Error())
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

	handlerExec.Command = handler.Command
	handlerExec.Timeout = int(handler.Timeout)
	handlerExec.Env = handler.EnvVars

	handlerExec.Input = string(eventData[:])

	result, err := command.ExecuteCommand(context.Background(), handlerExec)

	if err != nil {
		logger.Error("pipelined failed to execute event pipe handler: ", err.Error())
	} else {
		logger.Infof("pipelined executed event pipe handler: status=%x output=%s", result.Status, result.Output)
	}

	return result, err
}

// socketHandler creates either a TCP or UDP client to write eventData
// to a socket. The provided handler Type determines the protocol.
func (p *Pipelined) socketHandler(handler *types.Handler, eventData []byte) (conn net.Conn, err error) {
	protocol := handler.Type
	host := handler.Socket.Host
	port := handler.Socket.Port
	timeout := handler.Timeout

	// If Timeout is not specified, use the default.
	if timeout == 0 {
		timeout = DefaultSocketTimeout
	}

	address := fmt.Sprintf("%s:%d", host, port)
	timeoutDuration := time.Duration(timeout) * time.Second

	conn, err = net.DialTimeout(protocol, address, timeoutDuration)
	if err != nil {
		return nil, err
	}
	defer func() {
		e := conn.Close()
		if err == nil {
			err = e
		}
	}()

	bytes, err := conn.Write(eventData)

	if err != nil {
		logger.Errorf("pipelined failed to execute event %s handler: %v", protocol, err.Error())
	} else {
		logger.Debugf("pipelined executed event %s handler: bytes=%v", protocol, bytes)
	}

	return conn, nil
}
