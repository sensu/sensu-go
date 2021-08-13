package pipeline

import (
	"context"
	"errors"
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

// RunEventPipelines loops through an event's pipelines and runs them.
func (p *Pipeline) RunEventPipelines(ctx context.Context, event *corev2.Event) error {
	eventPipelines := []*corev2.Pipeline{}

	// convert the event's handlers to their own pipeline and add the pipeline
	// to the list of pipelines.
	if event.HasCheck() {

	}
	if event.HasMetrics() {

	}
	legacyCheckEventPipeline, err := p.pipelineFromEventHandlers(ctx, event)
	if err != nil {
		if _, ok := err.(*store.ErrInternal); ok {
			select {
			case p.errChan <- err:
			case <-p.stopping:
			}
			return
		}
		logger.Error(err)
	}
	pipelines = append(eventPipelines, legacyEventPipeline)

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

// pipelineFromHandlers converts an event's non-pipeline handlers
// (event.Handlers) and converts them to pipeline workflows under
// a single "LegacyPipeline".
func (p *Pipeline) pipelineFromHandlers(ctx context.Context, event *corev2.Event) (*corev2.Pipeline, error) {
	fields := logrus.Fields{
		"namespace": corev2.ContextNamespace(ctx),
	}

	handlers := []*corev2.Handler{}
	for _, handlerName := range event.Check.Handlers {
		fields["handler"] = handlerName

		tctx, cancel := context.WithTimeout(ctx, p.storeTimeout)
		handler, err := p.store.GetHandlerByName(tctx, handlerName)
		cancel()
		if err != nil {
			if _, ok := err.(*store.ErrInternal); ok {
				// fatal error
				return nil, err
			}
			logger.WithFields(fields).WithError(err).Error("failed to fetch handler")
			continue
		}
		if handler == nil {
			logger.WithFields(fields).WithError(err).Error("fetched handler is nil")
			continue
		}

		handlers = append(handlers, handler)
	}

	return corev2.PipelineFromHandlers(ctx, handlers), nil
}

type HandlerMap map[string][]*corev2.Handler

// expandHandlers turns a list of Sensu handler names into a list of
// handlers, while expanding handler sets with support for some
// nesting. Handlers are fetched from etcd.
func (p *Pipeline) expandHandlers(ctx context.Context, handlers []string, level int) (HandlerMap, error) {
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
		var extension *corev2.Extension

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
		} else {
			if _, ok := expanded[handler.Name]; !ok {
				expanded[handler.Name] = handlerExtensionUnion{Handler: handler, Extension: extension}
			}
		}
	}

	return expanded, nil
}
