package pipeline

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

type MutatorAdapter interface {
	Name() string
	CanMutate(*corev2.ResourceReference) bool
	Mutate(context.Context, *corev2.ResourceReference, *corev2.Event) ([]byte, error)
}

func (a *AdapterV1) processMutator(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) ([]byte, error) {
	mutator, err := a.getMutatorForResource(ctx, ref)
	if err != nil {
		return nil, err
	}

	return mutator.Mutate(ctx, ref, event)
}

func (a *AdapterV1) getMutatorForResource(ctx context.Context, ref *corev2.ResourceReference) (MutatorAdapter, error) {
	for _, mutatorAdapter := range a.MutatorAdapters {
		if mutatorAdapter.CanMutate(ref) {
			return mutatorAdapter, nil
		}
	}
	return nil, fmt.Errorf("no mutator adapters were found that can mutate the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}
