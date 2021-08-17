package pipeline

import (
	"context"
	"errors"
	"fmt"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/licensing"
	"github.com/sensu/sensu-go/backend/pipeline/legacy"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	utillogging "github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

const (
	LegacyEventPipelineName         = "LegacyEventPipeline"
	LegacyEventPipelineWorkflowName = "LegacyEventPipelineWorkflow-%s"
)

// Pipeline takes events as inputs, and treats them in various ways according
// to the event's check configuration.
type Pipeline struct {
	store                  store.Store
	assetGetter            asset.Getter
	backendEntity          *corev2.Entity
	executor               command.Executor
	storeTimeout           time.Duration
	secretsProviderManager *secrets.ProviderManager
	licenseGetter          licensing.Getter
	filters                []Filter
	mutators               []Mutator
	handlers               []Handler
}

func (p *Pipeline) AddFilter(filter Filter) {
	p.filters = append(p.filters, filter)
}

func (p *Pipeline) AddMutator(mutator Mutator) {
	p.mutators = append(p.mutators, mutator)
}

func (p *Pipeline) AddHandler(handler HandlerProcessor) {
	p.handlers = append(p.handlers, handler)
}

// DEPRECATED: use RunEventPipelines instead.
// HandleEvent takes a Sensu event through a Sensu pipeline, filters
// -> mutator -> handler. An event may have one or more handlers. Most
// errors are only logged and used for flow control, they will not
// interupt event handling.
func (p *Pipeline) HandleEvent(ctx context.Context, event *corev2.Event) error {
	return p.RunEventPipelines(ctx, event)
}

// RunEventPipelines loops through an event's pipelines and runs them.
func (p *Pipeline) RunEventPipelines(ctx context.Context, event *corev2.Event) error {
	ctx = context.WithValue(ctx, corev2.NamespaceKey, event.Entity.Namespace)

	// Prepare debug log entry
	debugFields := utillogging.EventFields(event, true)
	logger.WithFields(debugFields).Debug("received event")

	// Prepare log entry
	fields := utillogging.EventFields(event, false)

	// Get the event's pipelines
	eventPipelines, err := p.getEventPipelines(ctx, event)
	if err != nil {
		return err
	}
	if len(eventPipelines) == 0 {
		logger.WithFields(fields).Info("no pipelines available")
		return nil
	}

	return nil
}

// getEventPipelines resolves any pipeline references for a given event,
// constructs a legacy pipeline for any check/metric handlers, and then returns
// a slice of the pipelines.
func (p *Pipeline) getEventPipelines(ctx context.Context, event *corev2.Event) ([]*corev2.Pipeline, error) {
	// Prepare log entry
	fields := utillogging.EventFields(event, false)

	pipelines, err := p.expandEventPipelines(ctx, event.Pipelines)
	if err != nil {
		return nil, err
	}

	// generate a pipeline for any check and/or metric handlers that exist in
	// the event and add it to the list of event pipelines.
	legacyPipeline, err := p.generateLegacyEventPipeline(ctx, event)
	if err != nil {
		return nil, err
	}
	if legacyEventPipeline != nil {
		pipelines = append(pipelines, legacyPipeline)
	} else {
		logger.WithFields(fields).Debug("no handlers available to generate a legacy pipeline")
	}

	return pipelines, nil
}

func (p *Pipeline) resolveEventPipelines(ctx context.Context, pipelines []*corev2.ResourceReference) ([]*corev2.Pipeline, error) {
	expandedPipelines := []*corev2.Pipeline{}

	for _, pipeline := range pipelineRefs {
		tctx, cancel := context.WithTimeout(ctx, p.storeTimeout)
		
		handler, err := p.store.GetPipelineByName(tctx, pipelineName)
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

	return pipelines, nil
}

func (p *Pipeline) generateLegacyEventPipeline(ctx context.Context, event *corev2.Event) (*corev2.Pipeline, error) {
	// initialize a list of handler names for storing the names of any legacy
	// check and/or metrics handlers.
	legacyHandlerNames := []string{}

	if event.HasCheck() {
		legacyHandlerNames = append(legacyHandlerNames, event.Check.Handlers...)
	}

	if event.HasMetrics() {
		legacyHandlerNames = append(legacyHandlerNames, event.Metrics.Handlers...)
	}

	handlers, err := p.expandHandlers(ctx, legacyHandlerNames, 1)
	if err != nil {
		return nil, err
	}

	pipeline := &corev2.Pipeline{
		Metadata: &corev2.ObjectMeta{
			Name:      LegacyEventPipelineName,
			Namespace: corev2.ContextNamespace(ctx),
		},
		Workflows: []*corev2.PipelineWorkflow{},
	}

	for handlerName, handler := range handlers {
		workflowName := fmt.Sprintf(LegacyEventPipelineWorkflowName, handlerName)
		workflow := corev2.PipelineWorkflowFromHandler(ctx, workflowName, handler)
		pipeline.Workflows = append(pipeline.Workflows, workflow)
	}

	return pipeline, nil
}

func (p *Pipeline) runEventWorkflow(ctx context.Context, workflow *corev2.PipelineWorkflow, event *corev2.Event) error {
	// Process the event through the workflow filters
	filtered, err := p.runWorkflowFilters(ctx, workflow.Filters, event)
	if err != nil {
		return err
	}
	if filtered {
		return nil
	}

	// Process the event through the workflow mutator
	mutatedEvent, err := p.processMutator(ctx, workflow.Mutator, event)
	if err != nil {
		return err
	}
	if mutatedEvent != nil {
		event = mutatedEvent
	}

	// Process the event through the workflow handler
	return p.processHandler(ctx, workflow.Handler, event)
}

// New creates a new Pipeline from the provided configuration.
func New(c Config, options ...Option) *Pipeline {
	// default pipeline filters to search through when searching for a pipeline
	// filter that supports a referenced event filter resource.
	defaultFilters := []Filter{
		&legacy.Filter{
			AssetGetter:  c.AssetGetter,
			Store:        c.Store,
			StoreTimeout: c.StoreTimeout,
		},
	}

	// default pipeline mutators to search through when searching for a pipeline
	// mutator that supports a referenced event mutator resource.
	defaultMutators := []Mutator{
		&legacy.Mutator{
			Store:        c.Store,
			StoreTimeout: c.StoreTimeout,
		},
	}

	// default pipeline handlers to search through when searching for a pipeline
	// handler that supports a referenced event handler resource.
	defaultHandlers := []Handler{
		&legacy.Handler{
			AssetGetter:            c.AssetGetter,
			Executor:               command.NewExecutor(),
			LicenseGetter:          c.LicenseGetter,
			SecretsProviderManager: c.SecretsProviderManager,
			Store:                  c.Store,
			StoreTimeout:           c.StoreTimeout,
		},
	}

	pipeline := &Pipeline{
		store:                  c.Store,
		assetGetter:            c.AssetGetter,
		backendEntity:          c.BackendEntity,
		executor:               command.NewExecutor(),
		storeTimeout:           c.StoreTimeout,
		secretsProviderManager: c.SecretsProviderManager,
		licenseGetter:          c.LicenseGetter,
		filters:                defaultFilters,
		mutators:               defaultMutators,
		handlers:               defaultHandlers,
	}
	for _, o := range options {
		o(pipeline)
	}
	return pipeline
}

// expandHandlers turns a list of Sensu handler names into a list of
// handlers, while expanding handler sets with support for some
// nesting. Handlers are fetched from etcd.
func (p *Pipeline) expandHandlers(ctx context.Context, handlers []string, level int) ([]*corev2.Handler, error) {
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
		tctx, cancel := context.WithTimeout(ctx, p.storeTimeout)
		handler, err := p.store.GetHandlerByName(tctx, handlerName)
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
			setHandlers, err := p.expandHandlers(ctx, handler.Handlers, level+1)
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

	return expanded, nil
}

// func (p *Pipeline) HandleEvent(ctx context.Context, event *corev2.Event) error {
// 	ctx = context.WithValue(ctx, corev2.NamespaceKey, event.Entity.Namespace)

// 	// Prepare debug log entry
// 	debugFields := utillogging.EventFields(event, true)
// 	logger.WithFields(debugFields).Debug("received event")

// 	// Prepare log entry
// 	fields := utillogging.EventFields(event, false)

// 	var handlerList []string

// 	if event.HasCheck() {
// 		handlerList = append(handlerList, event.Check.Handlers...)
// 	}

// 	if event.HasMetrics() {
// 		handlerList = append(handlerList, event.Metrics.Handlers...)
// 	}

// 	handlers, err := p.expandHandlers(ctx, handlerList, 1)
// 	if err != nil {
// 		return err
// 	}

// 	if len(handlers) == 0 {
// 		logger.WithFields(fields).Info("no handlers available")
// 		return nil
// 	}

// 	for _, u := range handlers {
// 		handler := u.Handler
// 		fields["handler"] = handler.Name

// 		filter, err := p.FilterEvent(handler, event)
// 		if err != nil {
// 			if _, ok := err.(*store.ErrInternal); ok {
// 				// Fatal error
// 				return err
// 			}
// 			logger.WithError(err).Warn("error filtering event")
// 		}
// 		if filter != "" {
// 			logger.WithFields(fields).Infof("event filtered by filter %q", filter)
// 			continue
// 		}

// 		eventData, err := p.mutateEvent(handler, event)
// 		if err != nil {
// 			logger.WithFields(fields).WithError(err).Error("error mutating event")
// 			if _, ok := err.(*store.ErrInternal); ok {
// 				// Fatal error
// 				return err
// 			}
// 			continue
// 		}

// 		logger.WithFields(fields).Info("sending event to handler")
// 	}

// 	return nil
// }
