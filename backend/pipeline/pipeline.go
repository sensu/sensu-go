package pipeline

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/pipeline/filter"
	"github.com/sensu/sensu-go/backend/pipeline/handler"
	"github.com/sensu/sensu-go/backend/pipeline/mutator"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	"github.com/sirupsen/logrus"
)

type HandlerMap map[string]*corev2.Handler

// HandleEvent takes a Sensu event through its own pipelines. An event may have
// one or more pipelines. Most errors are only logged and used for flow control,
// they will not interupt event handling.
func (p *Pipeline) HandleEvent(ctx context.Context, event *corev2.Event) error {

}

func (p *Pipeline) getRunnerForResource(ctx context.Context, ref *corev2.ResourceReference) (PipelineAdapter, error) {
	for _, pipelineAdapter := range p.pipelineAdapters {
		if pipelineAdapter.CanRun(ctx, ref) {
			return pipelineAdapter, nil
		}
	}
	return nil, fmt.Errorf("no pipeline adapters were found that support the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}

// resolvePipelineReference fetches a core/v2.Pipeline reference from the
// store and returns a core/v2.Pipeline.
func (p *Pipeline) resolvePipelineReference(ctx context.Context, ref *corev2.ResourceReference) (*corev2.Pipeline, error) {
	// Prepare log entry
	fields := logrus.Fields{}

	pipelines := []*corev2.Pipeline{}

	for _, ref := range refs {
		
		}
	}

	return pipelines, nil
}

func (p *V1PipelineAdapter) CanRun(ctx context.Context, ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "Pipeline" {
		return true
	}
	return false
}

func (p *V1PipelineAdapter) Run(ctx context.Context, workflow *corev2.PipelineWorkflow, event *corev2.Event) error {
	// TODO: Either check for LegacyPipelineName here and determine whether or not to
	// call generateLegacyPipeline() or resolvePipelineReference(), or create
	// two different runners that separate the logic by adding a check for
	// the ref.Name equalling the value of LegacyPipelineName.

	
}

// New creates a new Pipeline from the provided configuration.
func New(c Config, options ...Option) *Pipeline {
	// default filter adapters to use within the pipeline
	defaultFilterAdapters := []FilterAdapter{
		&filter.Legacy{
			AssetGetter:  c.AssetGetter,
			Store:        c.Store,
			StoreTimeout: c.StoreTimeout,
		},
	}

	// default mutator adapters to use within the pipeline
	defaultMutatorAdapters := []MutatorAdapter{
		&mutator.Legacy{
			Store:        c.Store,
			StoreTimeout: c.StoreTimeout,
		},
	}

	// default handler adapters to use within the pipeline
	defaultHandlerAdapters := []HandlerAdapter{
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
		filterAdapters:         defaultFilterAdapters,
		mutatorAdapters:        defaultMutatorAdapters,
		handlerAdapters:        defaultHandlerAdapters,
	}
	for _, o := range options {
		o(pipeline)
	}
	return pipeline
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
