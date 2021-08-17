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

// RunEventPipelines loops through an event's pipelines and runs them.
func (p *Pipeline) RunEventPipelines(ctx context.Context, event *corev2.Event) error {
	eventPipelines := event.Pipelines

	// generate a pipeline for any check and/or metric handlers that exist in
	// the event and add it to the list of event pipelines.
	legacyEventPipeline, err := p.generateLegacyEventPipeline(ctx, event)
	if err != nil {
		return err
	}
	eventPipelines = append(eventPipelines, legacyEventPipeline)

	return nil
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
