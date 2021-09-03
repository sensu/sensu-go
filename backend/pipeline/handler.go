package pipeline

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

type HandlerAdapter interface {
	Name() string
	CanHandle(*corev2.ResourceReference) bool
	Handle(context.Context, *corev2.ResourceReference, *corev2.Event, []byte) error
}

func (a *AdapterV1) processHandler(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event, mutatedData []byte) error {
	handler, err := a.getHandlerForResource(ctx, ref)
	if err != nil {
		return err
	}

	return handler.Handle(ctx, ref, event, mutatedData)
}

func (a *AdapterV1) getHandlerForResource(ctx context.Context, ref *corev2.ResourceReference) (HandlerAdapter, error) {
	for _, handlerAdapter := range a.HandlerAdapters {
		if handlerAdapter.CanHandle(ref) {
			return handlerAdapter, nil
		}
	}
	return nil, fmt.Errorf("no handler adapters were found that can handle the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}
