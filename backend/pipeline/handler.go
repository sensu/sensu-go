package pipeline

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/rpc"
	"github.com/sensu/sensu-go/util/environment"
	utillogging "github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

type handlerExtensionUnion struct {
	*corev2.Extension
	*corev2.Handler
}

// HandleEvent takes a Sensu event through a Sensu pipeline, filters
// -> mutator -> handler. An event may have one or more handlers. Most
// errors are only logged and used for flow control, they will not
// interupt event handling.
func (p *Pipeline) HandleEvent(ctx context.Context, event *corev2.Event) error {
	ctx = context.WithValue(ctx, corev2.NamespaceKey, event.Entity.Namespace)

	// Prepare debug log entry
	debugFields := utillogging.EventFields(event, true)
	logger.WithFields(debugFields).Debug("received event")

	// Prepare log entry
	fields := utillogging.EventFields(event, false)

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

	if len(handlers) == 0 {
		logger.WithFields(fields).Info("no handlers available")
		return nil
	}

	for _, u := range handlers {
		handler := u.Handler
		fields["handler"] = handler.Name

		filtered, err := p.filterEvent(handler, event)
		if err != nil {
			if _, ok := err.(*store.ErrInternal); ok {
				// Fatal error
				return err
			}
			logger.WithError(err).Warn("error filtering event")
		}
		if filtered {
			logger.WithFields(fields).Info("event filtered")
			continue
		}

		eventData, err := p.mutateEvent(handler, event)
		if err != nil {
			logger.WithError(err).Warn("error mutating event")
			if _, ok := err.(*store.ErrInternal); ok {
				// Fatal error
				return err
			}
			continue
		}

		logger.WithFields(fields).Info("sending event to handler")

		switch handler.Type {
		case "pipe":
			if _, err := p.pipeHandler(handler, eventData); err != nil {
				logger.WithFields(fields).Error(err)
				if _, ok := err.(*store.ErrInternal); ok {
					return err
				}
			}
		case "tcp", "udp":
			if _, err := p.socketHandler(handler, eventData); err != nil {
				logger.WithFields(fields).Error(err)
				if _, ok := err.(*store.ErrInternal); ok {
					return err
				}
			}
		case "grpc":
			if _, err := p.grpcHandler(u.Extension, event, eventData); err != nil {
				logger.WithFields(fields).Error(err)
				if _, ok := err.(*store.ErrInternal); ok {
					return err
				}
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
func (p *Pipeline) expandHandlers(ctx context.Context, handlers []string, level int) (map[string]handlerExtensionUnion, error) {
	if level > 3 {
		return nil, errors.New("handler sets cannot be deeply nested")
	}

	expanded := map[string]handlerExtensionUnion{}

	// Prepare log entry
	namespace := corev2.ContextNamespace(ctx)
	fields := logrus.Fields{
		"namespace": namespace,
	}

	for _, handlerName := range handlers {
		tctx, cancel := context.WithTimeout(ctx, p.storeTimeout)
		handler, err := p.store.GetHandlerByName(tctx, handlerName)
		cancel()
		var extension *corev2.Extension

		// Add handler name to log entry
		fields["handler"] = handlerName

		if handler == nil {
			if err != nil {
				(logger.
					WithFields(fields).
					WithError(err).
					Error("failed to retrieve a handler"))
				if _, ok := err.(*store.ErrInternal); ok {
					// Fatal error
					return nil, err
				}
				continue
			}

			continue // remove this line if you enable the stuff below

			// TODO: this code enables extension handler lookups, but for now,
			// extensions are not enabled. Re-enable this code when extensions
			// are re-enabled.
			// extension, err = p.store.GetExtension(ctx, handlerName)
			// if err == store.ErrNoExtension {
			// 	continue
			// }
			// if err != nil {
			// 	(logger.
			// 		WithFields(fields).
			// 		WithError(err).
			// 		Error("failed to retrieve an extension"))
			// 	continue
			// }
			// handler = &corev2.Handler{
			// 	ObjectMeta: corev2.ObjectMeta{
			// 		Name: extension.URL,
			// 	},
			// 	Type: "grpc",
			// }
		}

		if handler.Type == "set" {
			setHandlers, err := p.expandHandlers(ctx, handler.Handlers, level+1)

			if err != nil {
				logger.
					WithFields(fields).
					WithError(err).
					Error("failed to expand handler set")
				if _, ok := err.(*store.ErrInternal); ok {
					return nil, err
				}
			} else {
				for name, u := range setHandlers {
					if _, ok := expanded[name]; !ok {
						expanded[name] = handlerExtensionUnion{Handler: u.Handler}
					}
				}
			}
		} else {
			if _, ok := expanded[handler.Name]; !ok {
				expanded[handler.Name] = handlerExtensionUnion{Handler: handler, Extension: extension}
			}
		}
	}

	return expanded, nil
}

// pipeHandler fork/executes a child process for a Sensu pipe handler
// command and writes the mutated eventData to it via STDIN.
func (p *Pipeline) pipeHandler(handler *corev2.Handler, eventData []byte) (*command.ExecutionResponse, error) {
	ctx := corev2.SetContextFromResource(context.Background(), handler)
	// Prepare log entry
	fields := logrus.Fields{
		"namespace": handler.Namespace,
		"handler":   handler.Name,
		"assets":    handler.RuntimeAssets,
	}

	secrets, err := p.secretsProviderManager.SubSecrets(ctx, handler.Secrets)
	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to retrieve secrets for handler")
		return nil, err
	}

	// Prepare environment variables
	env := environment.MergeEnvironments(os.Environ(), handler.EnvVars, secrets)

	handlerExec := command.ExecutionRequest{}
	handlerExec.Command = handler.Command
	handlerExec.Timeout = int(handler.Timeout)
	handlerExec.Env = env
	handlerExec.Input = string(eventData[:])

	// Only add assets to execution context if handler requires them
	if len(handler.RuntimeAssets) != 0 {
		logger.WithFields(fields).Debug("fetching assets for handler")
		// Fetch and install all assets required for handler execution
		matchedAssets := asset.GetAssets(ctx, p.store, handler.RuntimeAssets)

		assets, err := asset.GetAll(context.TODO(), p.assetGetter, matchedAssets)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("failed to retrieve assets for handler")
			if _, ok := err.(*store.ErrInternal); ok {
				// Fatal error
				return nil, err
			}
		} else {
			handlerExec.Env = environment.MergeEnvironments(os.Environ(), assets.Env(), handler.EnvVars, secrets)
		}
	}

	result, err := p.executor.Execute(context.Background(), handlerExec)

	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to execute event pipe handler")
	} else {
		fields["status"] = result.Status
		fields["output"] = result.Output
		logger.WithFields(fields).Info("event pipe handler executed")
	}

	return result, err
}

func (p *Pipeline) grpcHandler(ext *corev2.Extension, evt *corev2.Event, mutated []byte) (rpc.HandleEventResponse, error) {
	// Prepare log entry
	fields := logrus.Fields{
		"namespace": ext.GetNamespace(),
		"handler":   ext.GetName(),
	}

	logger.WithFields(fields).Debug("sending event to handler extension")

	executor, err := p.extensionExecutor(ext)
	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to execute event handler extension")
		return rpc.HandleEventResponse{}, err
	}
	defer func() {
		if err := executor.Close(); err != nil {
			logger.WithError(err).Debug("error closing grpc client conn")
		}
	}()

	result, err := executor.HandleEvent(evt, mutated)
	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to execute event handler extension")
	} else {
		fields["output"] = result.Output
		logger.WithFields(fields).Info("event handler extension executed")
	}

	return result, err
}

// socketHandler creates either a TCP or UDP client to write eventData
// to a socket. The provided handler Type determines the protocol.
func (p *Pipeline) socketHandler(handler *corev2.Handler, eventData []byte) (conn net.Conn, err error) {
	protocol := handler.Type
	host := handler.Socket.Host
	port := handler.Socket.Port
	timeout := handler.Timeout

	// Prepare log entry
	fields := logrus.Fields{
		"namespace": handler.Namespace,
		"handler":   handler.Name,
		"protocol":  protocol,
	}

	// If Timeout is not specified, use the default.
	if timeout == 0 {
		timeout = DefaultSocketTimeout
	}

	address := fmt.Sprintf("%s:%d", host, port)
	timeoutDuration := time.Duration(timeout) * time.Second

	logger.WithFields(fields).Debug("sending event to socket handler")

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
		logger.WithFields(fields).WithError(err).Error("failed to execute event handler")
	} else {
		fields["bytes"] = bytes
		logger.WithFields(fields).Info("event socket handler executed")
	}

	return conn, nil
}
