package pipeline

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	utillogging "github.com/sensu/sensu-go/util/logging"
)

type Handler interface {
	CanHandle(context.Context, *corev2.ResourceReference) bool
	Handle(context.Context, *corev2.ResourceReference, *corev2.Event) error
}

func (p *Pipeline) getHandlerForResource(ctx context.Context, ref *corev2.ResourceReference) (Handler, error) {
	for _, handler := range p.handlers {
		if handler.CanHandle(ctx, ref) {
			return handler, nil
		}
	}
	return nil, fmt.Errorf("no handler processors were found that can handle the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}

func (p *Pipeline) processHandler(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) error {
	handler, err := p.getHandlerForResource(ctx, ref)
	if err != nil {
		return err
	}

	return handler.Handle(ctx, ref, event)
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
