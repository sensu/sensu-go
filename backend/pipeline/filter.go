package pipeline

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

type Filter interface {
	Name() string
	CanFilter(context.Context, *corev2.ResourceReference) bool
	Filter(context.Context, *corev2.ResourceReference, *corev2.Event) (bool, error)
}

func (p *Pipeline) getFilterForResource(ctx context.Context, ref *corev2.ResourceReference) (Filter, error) {
	for _, filter := range p.filters {
		if filter.CanFilter(ctx, ref) {
			return filter, nil
		}
	}
	return nil, fmt.Errorf("no pipeline filters were found that can filter using resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}

func (p *Pipeline) processFilters(ctx context.Context, refs []*corev2.ResourceReference, event *corev2.Event) (bool, error) {
	// for each filter in the workflow, loop through each pipeline filter
	// until one is found that supports filtering the event using the referenced
	// resource.
	for _, ref := range refs {
		filter, err := p.getFilterForResource(ctx, ref)
		if err != nil {
			return false, err
		}

		filtered, err := filter.Filter(ctx, ref, event)
		if err != nil {
			return false, err
		}
		if filtered {
			return true, nil
		}
	}

	return false, nil
}
