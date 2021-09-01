package pipeline

import (
	"context"
	"errors"
	"fmt"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	utillogging "github.com/sensu/sensu-go/util/logging"
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
	return "V1Adapter"
}

func (a *AdapterV1) CanRun(ctx context.Context, ref *corev2.ResourceReference) bool {
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

	ctx = context.WithValue(ctx, corev2.NamespaceKey, event.Entity.Namespace)

	pipeline, err := a.resolvePipelineReference(ctx, ref, event)
	if err != nil {
		return err
	}

	if len(pipeline.Workflows) < 1 {
		logger.Info("pipeline has no workflows, skipping execution")
		return nil
	}

	for _, workflow := range pipeline.Workflows {
		// Process the event through the workflow filters
		filtered, err := a.processFilters(ctx, workflow.Filters, event)
		if err != nil {
			return err
		}
		if filtered {
			return nil
		}

		// Process the event through the workflow mutator
		mutatedData, err := a.processMutator(ctx, workflow.Mutator, event)
		if err != nil {
			return err
		}

		// Process the event through the workflow handler
		return a.processHandler(ctx, workflow.Handler, event, mutatedData)
	}

	// Prepare debug log entry
	debugFields := utillogging.EventFields(event, true)
	logger.WithFields(debugFields).Debug("received event")

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
	pipeline, err := a.Store.GetPipelineByName(tctx, ref.Name)
	cancel()

	if pipeline == nil {
		if err != nil {
			logger.
				WithFields(fields).
				WithError(err).
				Error("failed to retrieve a pipeline")
			if _, ok := err.(*store.ErrInternal); ok {
				// Fatal error
				return nil, err
			}
			return nil, nil
		}

		logger.WithFields(fields).Info("pipeline does not exist, will be ignored")
		return nil, nil
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
		Metadata: &corev2.ObjectMeta{
			Name:      LegacyPipelineName,
			Namespace: corev2.ContextNamespace(ctx),
		},
		Workflows: []*corev2.PipelineWorkflow{},
	}

	for handlerName, handler := range handlers {
		workflowName := fmt.Sprintf(LegacyPipelineWorkflowName, handlerName)
		workflow := corev2.PipelineWorkflowFromHandler(ctx, workflowName, handler)
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
			} else {
				for name, expandedHandler := range setHandlers {
					if _, ok := expandedHandlers[name]; !ok {
						expandedHandlers[name] = expandedHandler
					}
				}
			}
		}
	}

	return expandedHandlers, nil
}
