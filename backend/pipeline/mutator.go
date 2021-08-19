package pipeline

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

type Mutator interface {
	Name() string
	CanMutate(context.Context, *corev2.ResourceReference) bool
	Mutate(context.Context, *corev2.ResourceReference, *corev2.Event) (*corev2.Event, error)
}

func (p *Pipeline) getMutatorForResource(ctx context.Context, ref *corev2.ResourceReference) (Mutator, error) {
	for _, mutator := range p.mutators {
		if mutator.CanMutate(ctx, ref) {
			return mutator, nil
		}
	}
	return nil, fmt.Errorf("no pipeline mutators were found that can mutate the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}

func (p *Pipeline) processMutator(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (*corev2.Event, error) {
	mutator, err := p.getMutatorForResource(ctx, ref)
	if err != nil {
		return nil, err
	}

	return mutator.Mutate(ctx, ref, event)
}
