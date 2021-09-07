package pipeline

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sirupsen/logrus"
)

const (
	LegacyPipelineName         = "legacy-pipeline"
	LegacyPipelineWorkflowName = "legacy-pipeline-workflow-%s"
)

type HandlerMap map[string]*corev2.Handler

// AdapterV1 is a pipeline adapter that can run a pipeline for corev2.Events.
type AdapterV1 struct {
	Store           store.Store
	StoreTimeout    time.Duration
	FilterAdapters  []FilterAdapter
	MutatorAdapters []MutatorAdapter
	HandlerAdapters []HandlerAdapter
}

func (a *AdapterV1) Name() string {
	return "AdapterV1"
}

func (a *AdapterV1) CanRun(ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "Pipeline" {
		return true
	}
	return false
}

func (a *AdapterV1) Run(ctx context.Context, ref *corev2.ResourceReference, resource interface{}) error {
	event, ok := resource.(*corev2.Event)
	if !ok {
		return fmt.Errorf("resource is not a corev2.Event")
	}

	// Prepare log entry
	fields := event.LogFields(false)
	fields["adapter_name"] = a.Name()
	fields["pipeline"] = ref.LogFields(false)

	// Prepare debug log entry
	debugFields := event.LogFields(true)
	debugFields["adapter_name"] = fields["adapter_name"]
	debugFields["pipeline"] = fields["pipeline"]
	logger.WithFields(debugFields).Debugf("adapter received event")

	ctx = context.WithValue(ctx, corev2.NamespaceKey, event.Entity.Namespace)

	pipeline, err := a.resolvePipelineReference(ctx, ref, event)
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, corev2.PipelineKey, pipeline.Name)

	if len(pipeline.Workflows) < 1 {
		return errors.New("pipeline has no workflows")
	}

	for _, workflow := range pipeline.Workflows {
		ctx = context.WithValue(ctx, corev2.PipelineWorkflowKey, workflow.Name)

		fields["pipeline_workflow"] = workflow.Name
		debugFields["pipeline_workflow"] = workflow.Name

		// Process the event through the workflow filters
		filtered, err := a.processFilters(ctx, workflow.Filters, event)
		if err != nil {
			return err
		}
		if filtered {
			return errors.New("event was filtered")
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
		mutatedData, err := a.processMutator(ctx, workflow.Mutator, event)
		if err != nil {
			return err
		}

		// Process the event through the workflow handler
		return a.processHandler(ctx, workflow.Handler, event, mutatedData)
	}

	return nil
}

func (a *AdapterV1) resolvePipelineReference(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (*corev2.Pipeline, error) {
	if ref.Name == LegacyPipelineName {
		return a.generateLegacyPipeline(ctx, event)
	} else {
		return a.getPipelineFromStore(ctx, ref)
	}
}

// getPipelineFromStore fetches a core/v2.Pipeline reference from the store and
// returns a core/v2.Pipeline.
func (a *AdapterV1) getPipelineFromStore(ctx context.Context, ref *corev2.ResourceReference) (*corev2.Pipeline, error) {
	tctx, cancel := context.WithTimeout(ctx, a.StoreTimeout)
	defer cancel()

	pipeline, err := a.Store.GetPipelineByName(tctx, ref.Name)
	if err != nil {
		return nil, err
	}
	if pipeline == nil {
		return nil, errors.New("pipeline does not exist")
	}

	return pipeline, nil
}

// generateLegacyPipeline will build an event pipeline with a pipeline
// workflow for each event.Check.Handlers & event.Metrics.Handlers
func (a *AdapterV1) generateLegacyPipeline(ctx context.Context, event *corev2.Event) (*corev2.Pipeline, error) {
	// initialize a list of handler names for storing the names of any legacy
	// check and/or metrics handlers.
	legacyHandlerNames := []string{}

	if event.HasCheck() {
		legacyHandlerNames = append(legacyHandlerNames, event.Check.Handlers...)
	}

	if event.HasMetrics() {
		legacyHandlerNames = append(legacyHandlerNames, event.Metrics.Handlers...)
	}

	handlers, err := a.expandHandlers(ctx, legacyHandlerNames, 1)
	if err != nil {
		return nil, err
	}

	pipeline := &corev2.Pipeline{
		ObjectMeta: corev2.ObjectMeta{
			Name:      LegacyPipelineName,
			Namespace: event.GetNamespace(),
		},
		Workflows: []*corev2.PipelineWorkflow{},
	}

	// sort the keys of the handlers map to guarantee the ordering of the
	// slice of handlers
	handlerNames := make([]string, 0, len(handlers))
	for handlerName := range handlers {
		handlerNames = append(handlerNames, handlerName)
	}
	sort.Strings(handlerNames)

	for _, handlerName := range handlerNames {
		workflowName := fmt.Sprintf(LegacyPipelineWorkflowName, handlerName)
		workflow := corev2.PipelineWorkflowFromHandler(ctx, workflowName, handlers[handlerName])
		pipeline.Workflows = append(pipeline.Workflows, workflow)
	}

	return pipeline, nil
}

// expandHandlers turns a list of Sensu handler names into a list of
// handlers, while expanding handler sets with support for some
// nesting. Handlers are fetched from etcd.
func (a *AdapterV1) expandHandlers(ctx context.Context, handlers []string, level int) (HandlerMap, error) {
	if level > 3 {
		return nil, errors.New("handler sets cannot be deeply nested")
	}

	expandedHandlers := HandlerMap{}

	// Prepare log entry
	namespace := corev2.ContextNamespace(ctx)
	fields := logrus.Fields{
		"namespace": namespace,
	}

	for _, handlerName := range handlers {
		tctx, cancel := context.WithTimeout(ctx, a.StoreTimeout)
		handler, err := a.Store.GetHandlerByName(tctx, handlerName)
		cancel()

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
			continue
		}

		if handler.Type == "set" {
			setHandlers, err := a.expandHandlers(ctx, handler.Handlers, level+1)
			if err != nil {
				logger.
					WithFields(fields).
					WithError(err).
					Error("failed to expand handler set")
				if _, ok := err.(*store.ErrInternal); ok {
					return nil, err
				}
				// TODO(jk): do we intend to continue here despite receiving
				// an error?
			} else {
				for name, expandedHandler := range setHandlers {
					if _, ok := expandedHandlers[name]; !ok {
						expandedHandlers[name] = expandedHandler
					}
				}
			}
		} else {
			expandedHandlers[handler.Name] = handler
		}
	}

	return expandedHandlers, nil
}
