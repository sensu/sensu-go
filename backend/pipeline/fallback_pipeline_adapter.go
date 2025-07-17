package pipeline

import (
	"context"
	"fmt"
	"time"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

type FallbackPipelinesAdapter struct {
	Store           store.Store
	StoreTimeout    time.Duration
	FilterAdapters  []FilterAdapter
	MutatorAdapters []MutatorAdapter
	HandlerAdapters []HandlerAdapter
}

// Name returns the name of the FallbackPipelinesAdapter
func (f *FallbackPipelinesAdapter) Name() string {
	return "FallbackPipelinesAdapter"
}

// FallbackPipelinesAdapter is an adapter for fallback pipelines
func (f *FallbackPipelinesAdapter) CanRun(ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "FallbackPipelines" {
		return true
	}
	return false
}

// Run executes the fallback pipelines for the given resource reference and event.
func (f *FallbackPipelinesAdapter) Run(ctx context.Context, ref *corev2.ResourceReference, resource interface{}) error {
	event, ok := resource.(*corev2.Event)
	if !ok {
		return fmt.Errorf("resource is not a corev2.Event")
	}

	// Prepare log entry
	fields := event.LogFields(false)
	fields["adapter_name"] = f.Name()
	fields["pipeline"] = ref.LogFields(false)

	// Prepare debug log entry
	debugFields := event.LogFields(true)
	debugFields["adapter_name"] = fields["adapter_name"]
	debugFields["pipeline"] = fields["pipeline"]
	logger.WithFields(debugFields).Debugf("adapter received event")

	ctx = context.WithValue(ctx, corev2.NamespaceKey, event.Entity.Namespace)

	pipelines, err := f.getFallbackPipelinesFromStore(ctx, ref, event)
	if err != nil {
		return err
	}

	if len(pipelines) < 1 {
		return &ErrNoFallbackPipelines{}
	}

	for _, pipeline := range pipelines {
		// TODO execute all Pipelines one by one until one succeeds or there are no more Pipelines to execute
		if err := f.executePipeline(ctx, pipeline, event, fields, debugFields); err != nil {
			// TODO log the error may be, do something according to proposal doc
			continue
		}
		return nil
	}

	return &ErrEndOfFallbackPipelines{}
}

func (f *FallbackPipelinesAdapter) executePipeline(ctx context.Context, pipeline *corev2.Pipeline, event *corev2.Event, fields, debugFields map[string]interface{}) error {
	for _, workflow := range pipeline.Workflows {
		ctx = context.WithValue(ctx, corev2.PipelineWorkflowKey, workflow.Name)

		fields["pipeline_workflow"] = workflow.Name
		debugFields["pipeline_workflow"] = workflow.Name

		// Process the event through the workflow filters
		filtered, err := f.processFilters(ctx, workflow.Filters, event)
		if err != nil {
			return err
		}
		if filtered {
			continue
		}

		// If no workflow mutator is set, use the JSON mutator
		if workflow.Mutator == nil {
			workflow.Mutator = &corev2.ResourceReference{
				APIVersion: "core/v2",
				Type:       "Mutator",
				Name:       "json",
			}
		}

		// Process the event through the workflow mutator
		mutatedData, err := f.processMutator(ctx, workflow.Mutator, event)
		if err != nil {
			return err
		}

		// Process the event through the workflow handler
		handlerRequestsTotalCounter.Inc()
		err = f.processHandler(ctx, workflow.Handler, event, mutatedData)
		incrementCounter(workflow.Handler, err)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FallbackPipelinesAdapter) processFilters(ctx context.Context, refs []*corev2.ResourceReference, event *corev2.Event) (bool, error) {
	// for each filter reference in the workflow, attempt to find a compatible
	// filter adapter and use it to filter the event.
	for _, ref := range refs {
		filtered, err := f.processFilter(ctx, ref, event)
		if err != nil {
			return false, err
		}
		if filtered {
			return true, nil
		}
	}
	return false, nil
}

func (f *FallbackPipelinesAdapter) processFilter(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (bool, error) {
	filter, err := f.getFilterAdapterForResource(ctx, ref)
	if err != nil {
		return false, err
	}

	return filter.Filter(ctx, ref, event)
}

func (f *FallbackPipelinesAdapter) getFilterAdapterForResource(ctx context.Context, ref *corev2.ResourceReference) (FilterAdapter, error) {
	for _, filterAdapter := range f.FilterAdapters {
		if filterAdapter.CanFilter(ref) {
			return filterAdapter, nil
		}
	}
	return nil, fmt.Errorf("no filter adapters were found that can filter the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}

func (f *FallbackPipelinesAdapter) processMutator(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) ([]byte, error) {
	mutator, err := f.getMutatorAdapterForResource(ctx, ref)
	if err != nil {
		return nil, err
	}

	return mutator.Mutate(ctx, ref, event)
}

func (f *FallbackPipelinesAdapter) getMutatorAdapterForResource(ctx context.Context, ref *corev2.ResourceReference) (MutatorAdapter, error) {
	for _, mutatorAdapter := range f.MutatorAdapters {
		if mutatorAdapter.CanMutate(ref) {
			return mutatorAdapter, nil
		}
	}
	return nil, fmt.Errorf("no mutator adapters were found that can mutate the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}

func (f *FallbackPipelinesAdapter) processHandler(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event, mutatedData []byte) error {
	handler, err := f.getHandlerAdapterForResource(ctx, ref)
	if err != nil {
		return err
	}

	return handler.Handle(ctx, ref, event, mutatedData)
}

func (f *FallbackPipelinesAdapter) getHandlerAdapterForResource(ctx context.Context, ref *corev2.ResourceReference) (HandlerAdapter, error) {
	for _, handlerAdapter := range f.HandlerAdapters {
		if handlerAdapter.CanHandle(ref) {
			return handlerAdapter, nil
		}
	}
	return nil, fmt.Errorf("no handler adapters were found that can handle the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}

// TODO - implement this method to retrieve fallback pipelines from the store
func (f *FallbackPipelinesAdapter) getFallbackPipelinesFromStore(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) ([]*corev2.Pipeline, error) {
	return []*corev2.Pipeline{
		{
			Workflows: []*corev2.PipelineWorkflow{},
		},
	}, nil
}
