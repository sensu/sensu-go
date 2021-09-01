package pipeline

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

type FilterAdapter interface {
	Name() string
	CanFilter(context.Context, *corev2.ResourceReference) bool
	Filter(context.Context, *corev2.ResourceReference, *corev2.Event) (bool, error)
}

func (a *AdapterV1) processFilters(ctx context.Context, refs []*corev2.ResourceReference, event *corev2.Event) (bool, error) {
	// for each filter in the workflow, loop through each pipeline filter
	// until one is found that supports filtering the event using the referenced
	// resource.
	for _, ref := range refs {
		filter, err := a.getFilterForResource(ctx, ref)
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

func (a *AdapterV1) getFilterForResource(ctx context.Context, ref *corev2.ResourceReference) (FilterAdapter, error) {
	for _, filterAdapter := range a.FilterAdapters {
		if filterAdapter.CanFilter(ctx, ref) {
			return filterAdapter, nil
		}
	}
	return nil, fmt.Errorf("no filter adapters were found that can filter the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}
