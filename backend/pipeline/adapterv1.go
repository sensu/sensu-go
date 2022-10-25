package pipeline

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	metricspkg "github.com/sensu/sensu-go/metrics"
	"github.com/sirupsen/logrus"
)

const (
	LegacyPipelineName         = "legacy-pipeline"
	LegacyPipelineWorkflowName = "legacy-pipeline-workflow-%s"

	HandlerRequests      = "sensu_go_handler_requests"
	HandlerRequestsTotal = "sensu_go_handler_requests_total"

	// PipelineDuration is the name of the prometheus summary vec used to track
	// average latencies of pipeline execution.
	PipelineDuration = "sensu_go_pipeline_duration"

	// PipelineResolveDuration is the name of the prometheus summary vec used to
	// track average latencies of pipeline reference resolving.
	PipelineResolveDuration = "sensu_go_pipeline_resolve_duration"

	// PipelineTypeLabelName is the name of a label which describes what type of
	// pipeline a metric is being recorded for.
	PipelineTypeLabelName = "pipeline_type"

	// PipelineTypeLabelLegacy is the value to use for the pipeline_type label
	// when the metric is for a legacy pipeline.
	PipelineTypeLabelLegacy = "legacy"

	// PipelineTypeLabelModern is the value to use for the pipeline_type label
	// when the metric is for a modern pipeline.
	PipelineTypeLabelModern = "modern"
)

var (
	handlerRequestsTotalCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: HandlerRequestsTotal,
			Help: "The total number of handler requests invoked",
		},
	)

	handlerRequestsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: HandlerRequests,
			Help: "The number of processed handler requests",
		},
		[]string{"status", "type"},
	)

	pipelineDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       PipelineDuration,
			Help:       "pipeline execution latency distribution",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{metricspkg.StatusLabelName, metricspkg.ResourceReferenceLabelName},
	)

	pipelineResolveDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       PipelineResolveDuration,
			Help:       "pipeline reference resolving latency distribution",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{metricspkg.StatusLabelName, PipelineTypeLabelName},
	)
)

func init() {
	legacyResourceID := LegacyPipelineReference().ResourceID()
	pipelineDuration.WithLabelValues(metricspkg.StatusLabelSuccess, legacyResourceID)
	pipelineDuration.WithLabelValues(metricspkg.StatusLabelError, legacyResourceID)

	pipelineResolveDuration.WithLabelValues(metricspkg.StatusLabelSuccess, PipelineTypeLabelLegacy)
	pipelineResolveDuration.WithLabelValues(metricspkg.StatusLabelSuccess, PipelineTypeLabelModern)
	pipelineResolveDuration.WithLabelValues(metricspkg.StatusLabelError, PipelineTypeLabelLegacy)
	pipelineResolveDuration.WithLabelValues(metricspkg.StatusLabelError, PipelineTypeLabelModern)

	if err := prometheus.Register(handlerRequestsTotalCounter); err != nil {
		panic(fmt.Errorf("error registering %s: %s", HandlerRequestsTotal, err))
	}
	if err := prometheus.Register(handlerRequestsCounter); err != nil {
		panic(fmt.Errorf("error registering %s: %s", HandlerRequests, err))
	}
	if err := prometheus.Register(pipelineDuration); err != nil {
		panic(fmt.Errorf("error registering %s: %s", PipelineDuration, err))
	}
	if err := prometheus.Register(pipelineResolveDuration); err != nil {
		panic(fmt.Errorf("error registering %s: %s", PipelineResolveDuration, err))
	}
}

func LegacyPipelineReference() *corev2.ResourceReference {
	return &corev2.ResourceReference{
		APIVersion: "core/v2",
		Type:       "LegacyPipeline",
		Name:       LegacyPipelineName,
	}
}

type HandlerMap map[string]*corev2.Handler

// AdapterV1 is a pipeline adapter that can run a pipeline for corev2.Events.
type AdapterV1 struct {
	Store           storev2.Interface
	StoreTimeout    time.Duration
	FilterAdapters  []FilterAdapter
	MutatorAdapters []MutatorAdapter
	HandlerAdapters []HandlerAdapter
}

func (a *AdapterV1) Name() string {
	return "AdapterV1"
}

func (a *AdapterV1) CanRun(ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" {
		if ref.Type == "Pipeline" || ref.Type == "LegacyPipeline" {
			return true
		}
	}
	return false
}

func (a *AdapterV1) Run(ctx context.Context, ref *corev2.ResourceReference, resource interface{}) (fErr error) {
	begin := time.Now()
	defer func() {
		duration := time.Since(begin)
		status := metricspkg.StatusLabelSuccess
		if fErr != nil {
			status = metricspkg.StatusLabelError
		}
		pipelineDuration.
			WithLabelValues(status, ref.ResourceID()).
			Observe(float64(duration) / float64(time.Millisecond))
	}()

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
		return &ErrNoWorkflows{}
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
		mutatedData, err := a.processMutator(ctx, workflow.Mutator, event)
		if err != nil {
			return err
		}

		// Process the event through the workflow handler
		handlerRequestsTotalCounter.Inc()
		err = a.processHandler(ctx, workflow.Handler, event, mutatedData)
		incrementCounter(workflow.Handler, err)
		if err != nil {
			return err
		}
	}

	return nil
}

func incrementCounter(handler *corev2.ResourceReference, err error) {
	handlerType := fmt.Sprintf("%s.%s", handler.GetAPIVersion(), handler.GetType())
	status := "0"
	if err != nil {
		status = "1"
	}
	handlerRequestsCounter.WithLabelValues(status, handlerType).Inc()
}

func (a *AdapterV1) resolvePipelineReference(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (pipeline *corev2.Pipeline, err error) {
	isLegacy := ref.Name == LegacyPipelineName

	resolveTimer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		status := metricspkg.StatusLabelSuccess
		if err != nil {
			status = metricspkg.StatusLabelError
		}
		pipelineType := PipelineTypeLabelModern
		if isLegacy {
			pipelineType = PipelineTypeLabelLegacy
		}
		pipelineResolveDuration.WithLabelValues(status, pipelineType).Observe(v * float64(1000))
	}))
	defer resolveTimer.ObserveDuration()

	if isLegacy {
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

	store := storev2.NewGenericStore[*corev2.Pipeline](a.Store)
	id := storev2.ID{Namespace: corev2.ContextNamespace(ctx), Name: ref.Name}
	return store.Get(tctx, id)
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

	handlers, err := a.expandHandlers(ctx, event.Entity.Namespace, legacyHandlerNames, 1)
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
func (a *AdapterV1) expandHandlers(ctx context.Context, namespace string, handlers []string, level int) (HandlerMap, error) {
	if level > 3 {
		return nil, errors.New("handler sets cannot be deeply nested")
	}

	expandedHandlers := HandlerMap{}

	// Prepare log entry
	fields := logrus.Fields{
		"namespace": namespace,
	}

	hstore := storev2.NewGenericStore[*corev2.Handler](a.Store)

	for _, handlerName := range handlers {
		tctx, cancel := context.WithTimeout(ctx, a.StoreTimeout)
		id := storev2.ID{Namespace: namespace, Name: handlerName}
		handler, err := hstore.Get(tctx, id)
		cancel()

		// Add handler name to log entry
		fields["handler"] = handlerName

		if err != nil {
			if _, ok := err.(*store.ErrNotFound); ok {
				if level > 1 {
					logger.WithFields(fields).Error("set handler specified a handler that does not exist")
				} else {
					logger.WithFields(fields).Info("handler does not exist, will be ignored")
				}
				continue
			}
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

		if handler.Type == "set" {
			setHandlers, err := a.expandHandlers(ctx, namespace, handler.Handlers, level+1)
			if err != nil {
				logger.
					WithFields(fields).
					WithError(err).
					Error("failed to expand handler set")
				if _, ok := err.(*store.ErrInternal); ok {
					return nil, err
				}
				// TODO(jk): do we intend to continue here despite receiving // nosemgrep:dgryski.semgrep-go.errtodo.err-todo
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
