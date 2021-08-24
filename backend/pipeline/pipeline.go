package pipeline

import (
	"context"
	"errors"
	"fmt"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/licensing"
	"github.com/sensu/sensu-go/backend/pipeline/filter"
	"github.com/sensu/sensu-go/backend/pipeline/handler"
	"github.com/sensu/sensu-go/backend/pipeline/mutator"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	utillogging "github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

const (
	LegacyPipelineName         = "legacy-pipeline"
	LegacyPipelineWorkflowName = "legacy-pipeline-workflow-%s"
)

type HandlerMap map[string]*corev2.Handler

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
	runners                []Runner
}

func (p *Pipeline) AddFilter(filter Filter) {
	p.filters = append(p.filters, filter)
}

func (p *Pipeline) AddMutator(mutator Mutator) {
	p.mutators = append(p.mutators, mutator)
}

func (p *Pipeline) AddHandler(handler Handler) {
	p.handlers = append(p.handlers, handler)
}

// HandleEvent takes a Sensu event through its own pipelines. An event may have
// one or more pipelines. Most errors are only logged and used for flow control,
// they will not interupt event handling.
func (p *Pipeline) HandleEvent(ctx context.Context, event *corev2.Event) error {
	ctx = context.WithValue(ctx, corev2.NamespaceKey, event.Entity.Namespace)

	// Prepare debug log entry
	debugFields := utillogging.EventFields(event, true)
	logger.WithFields(debugFields).Debug("received event")

	// Prepare log entry
	fields := utillogging.EventFields(event, false)

	// Construct a list of pipeline references
	pipelineRefs := event.Pipelines

	// Add a legacy pipeline "reference" if the event has handlers
	if event.HasHandlers() {
		legacyPipelineRef := &corev2.ResourceReference{
			APIVersion: "core/v2",
			Type:       "Pipeline",
			Name:       LegacyPipelineName,
		}
		pipelineRefs = append(pipelineRefs, legacyPipelineRef)
	} else {
		logger.WithFields(fields).Debug("no handlers available to generate a legacy pipeline")
	}

	if len(pipelineRefs) == 0 {
		logger.WithFields(fields).Info("no pipelines defined")
		return nil
	}

	// Loop through each event pipeline, search for a compatible pipeline
	// runner & run it
	for _, pipelineRef := range pipelineRefs {
		for _, runner := range p.runners {
			if runner.CanRun(ctx, pipelineRef) {
				runner.Run(ctx, pipelineRef, event)
			}
		}
	}

	return nil
}

// A runner will resolve and run a pipeline reference so long as CanRun()
// returns true for the ResourceReference.
type Runner interface {
	Name() string
	CanRun(context.Context, *corev2.ResourceReference) bool
	Run(context.Context, *corev2.ResourceReference, interface{}) error
}

type StandardPipelineRunner struct{}

func (p *Pipeline) getRunnerForResource(ctx context.Context, ref *corev2.ResourceReference) (PipelineRunner, error) {
	for _, runner := range p.runners {
		if runner.CanRun(ctx, ref) {
			return runner, nil
		}
	}
	return nil, fmt.Errorf("no pipeline runners were found that support resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}

// generateLegacyPipeline will build an event pipeline with a pipeline
// workflow for each event.Check.Handlers & event.Metrics.Handlers
func (p *Pipeline) generateLegacyPipeline(ctx context.Context, event *corev2.Event) (*corev2.Pipeline, error) {
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

// resolvePipelineReference fetches a core/v2.Pipeline reference from the
// store and returns a core/v2.Pipeline.
func (p *Pipeline) resolvePipelineReference(ctx context.Context, ref *corev2.ResourceReference) (*corev2.Pipeline, error) {
	// Prepare log entry
	fields := logrus.Fields{}

	pipelines := []*corev2.Pipeline{}

	for _, ref := range refs {
		fields["pipeline_reference"] = ref.ResourceID()

		// TODO: introduce a pipeline.Executor interface type and a
		// pipelineExecutors field to Pipeline. Loop through each
		// pipelineExecutor to find one that can support each pipeline
		// reference.
		if !p.canExecuteEventPipeline(ctx, ref) {
			logger.WithFields(fields).Info("no pipeline executors support pipeline reference, will be ignored")
		}

		tctx, cancel := context.WithTimeout(ctx, p.storeTimeout)
		pipeline, err := p.store.GetPipelineByName(tctx, ref.Name)
		cancel()

		if pipeline == nil {
			if err != nil {
				(logger.
					WithFields(fields).
					WithError(err).
					Error("failed to retrieve a pipeline"))
				if _, ok := err.(*store.ErrInternal); ok {
					// Fatal error
					return nil, err
				}
				continue
			}

			logger.WithFields(fields).Info("pipeline does not exist, will be ignored")
			continue
		}
	}

	return pipelines, nil
}

func (p *StandardPipelineRunner) CanRun(ctx context.Context, ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "Pipeline" {
		return true
	}
	return false
}

func (p *StandardPipelineRunner) Run(ctx context.Context, workflow *corev2.PipelineWorkflow, event *corev2.Event) error {
	// TODO: Either check for LegacyPipelineName here and determine whether or not to
	// call generateLegacyPipeline() or resolvePipelineReference(), or create
	// two different runners that separate the logic by adding a check for
	// the ref.Name equalling the value of LegacyPipelineName.

	// Process the event through the workflow filters
	filtered, err := p.processFilters(ctx, workflow.Filters, event)
	if err != nil {
		return err
	}
	if filtered {
		return nil
	}

	// Process the event through the workflow mutator
	mutatedData, err := p.processMutator(ctx, workflow.Mutator, event)
	if err != nil {
		return err
	}

	// Process the event through the workflow handler
	return p.processHandler(ctx, workflow.Handler, event, mutatedData)
}

// New creates a new Pipeline from the provided configuration.
func New(c Config, options ...Option) *Pipeline {
	// default pipeline filters to search through when searching for a pipeline
	// filter that supports a referenced event filter resource.
	defaultFilters := []Filter{
		&filter.Legacy{
			AssetGetter:  c.AssetGetter,
			Store:        c.Store,
			StoreTimeout: c.StoreTimeout,
		},
	}

	// default pipeline mutators to search through when searching for a pipeline
	// mutator that supports a referenced event mutator resource.
	defaultMutators := []Mutator{
		&mutator.Legacy{
			Store:        c.Store,
			StoreTimeout: c.StoreTimeout,
		},
	}

	// default pipeline handlers to search through when searching for a pipeline
	// handler that supports a referenced event handler resource.
	defaultHandlers := []Handler{
		&handler.Legacy{
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

	return expandedHandlers, nil
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
