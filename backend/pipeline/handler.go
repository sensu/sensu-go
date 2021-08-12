package pipeline

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	utillogging "github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

type Handler interface {
	CanHandle(context.Context, *corev2.ResourceReference) bool
	Handle(context.Context, *corev2.ResourceReference, *corev2.Event) error
}

func (p *Pipeline) getFilterProcessorForResource(ctx context.Context, ref *corev3.ResourceReference) (Filter, error) {
	for _, processor := range p.filters {
		if processor.CanFilter(ctx, ref) {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("no filter processors were found that can filter the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}

func (p *Pipeline) getMutatorProcessorForResource(ctx context.Context, ref *corev3.ResourceReference) (Mutator, error) {
	for _, processor := range p.mutators {
		if processor.CanMutate(ctx, ref) {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("no mutator processors were found that can mutate the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}

func (p *Pipeline) getHandlerProcessorForResource(ctx context.Context, ref *corev3.ResourceReference) (Handler, error) {
	for _, processor := range p.handlers {
		if processor.CanHandle(ctx, ref) {
			return processor, nil
		}
	}
	return nil, fmt.Errorf("no handler processors were found that can handle the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}

func (p *Pipeline) processFilters(ctx context.Context, refs []*corev3.ResourceReference, event *corev2.Event) (bool, error) {
	// for each filter in the workflow, loop through each filter processor
	// until one is found that supports filtering the event using the referenced
	// resource.
	for _, ref := range refs {
		processor, err := p.getFilterProcessorForResource(ctx, ref)
		if err != nil {
			return false, err
		}

		filtered, err := processor.Filter(ctx, ref, event)
		if err != nil {
			return false, err
		}
		if filtered {
			return true, nil
		}
	}

	return false, nil
}

func (p *Pipeline) processMutator(ctx context.Context, ref *corev3.ResourceReference, event *corev2.Event) (*corev2.Event, error) {
	processor, err := p.getMutatorProcessorForResource(ctx, ref)
	if err != nil {
		return nil, err
	}

	return processor.Mutate(ctx, ref, event)
}

func (p *Pipeline) processHandler(ctx context.Context, ref *corev3.ResourceReference, event *corev2.Event) error {
	processor, err := p.getHandlerProcessorForResource(ctx, ref)
	if err != nil {
		return err
	}

	return processor.Handle(ctx, ref, event)
}

func (p *Pipeline) ExecuteWorkflowForEvent(ctx context.Context, workflow *corev3.PipelineWorkflow, event *corev2.Event) error {
	// Process the event through the workflow filters
	filtered, err := p.processFilters(ctx, workflow.Filters, event)
	if err != nil {
		return err
	}
	if filtered {
		return nil
	}

	// Process the event through the workflow mutator
	mutatedEvent, err := p.processMutator(ctx, workflow.Mutator, event)
	if err != nil {
		return err
	}
	if mutatedEvent != nil {
		event = mutatedEvent
	}

	// Process the event through the workflow handler
	return p.processHandler(ctx, workflow.Handler, event)
}

// RunEventPipelines loops through an event's pipelines and runs them.
func (p *Pipeline) RunEventPipelines(ctx context.Context, event *corev2.Event) error {
	eventPipelines := []*corev2.Pipeline{}

	// convert the event's handlers to their own pipeline and add the pipeline
	// to the list of pipelines.
	if event.HasCheck() {

	}
	if event.HasMetrics() {

	}
	legacyCheckEventPipeline, err := p.pipelineFromEventHandlers(ctx, event)
	if err != nil {
		if _, ok := err.(*store.ErrInternal); ok {
			select {
			case p.errChan <- err:
			case <-p.stopping:
			}
			return
		}
		logger.Error(err)
	}
	pipelines = append(pipelines, legacyEventPipeline)

	return nil
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

		filter, err := p.FilterEvent(handler, event)
		if err != nil {
			if _, ok := err.(*store.ErrInternal); ok {
				// Fatal error
				return err
			}
			logger.WithError(err).Warn("error filtering event")
		}
		if filter != "" {
			logger.WithFields(fields).Infof("event filtered by filter %q", filter)
			continue
		}

		eventData, err := p.mutateEvent(handler, event)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("error mutating event")
			if _, ok := err.(*store.ErrInternal); ok {
				// Fatal error
				return err
			}
			continue
		}

		logger.WithFields(fields).Info("sending event to handler")
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

			logger.WithFields(fields).Info("handler does not exist, will be ignored")
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

// socketHandler creates either a TCP or UDP client to write eventData
// to a socket. The provided handler Type determines the protocol.
func (p *Pipeline) socketHandler(handler *corev2.Handler, event *corev2.Event, eventData []byte) (conn net.Conn, err error) {
	protocol := handler.Type
	host := handler.Socket.Host
	port := handler.Socket.Port
	timeout := handler.Timeout

	// Prepare log entry
	fields := utillogging.EventFields(event, false)
	fields["handler_name"] = handler.Name
	fields["handler_namespace"] = handler.Namespace
	fields["handler_protocol"] = protocol

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
