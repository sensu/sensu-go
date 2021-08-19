package pipeline

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

type Handler interface {
	Name() string
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
