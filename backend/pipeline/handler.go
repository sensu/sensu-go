package pipeline

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	metricspkg "github.com/sensu/sensu-go/metrics"
)

const (
	// HandlerDuration is the name of the prometheus summary vec used to track
	// average latencies of pipeline handler execution.
	HandlerDuration = "sensu_go_pipeline_handler_duration"
)

var (
	handlerDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       HandlerDuration,
			Help:       "pipeline handler execution latency distribution",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{metricspkg.StatusLabelName, metricspkg.ResourceReferenceLabelName},
	)
)

type HandlerAdapter interface {
	Name() string
	CanHandle(*corev2.ResourceReference) bool
	Handle(context.Context, *corev2.ResourceReference, *corev2.Event, []byte) error
}

func init() {
	if err := prometheus.Register(handlerDuration); err != nil {
		panic(fmt.Errorf("error registering %s: %s", HandlerDuration, err))
	}
}

func (a *AdapterV1) processHandler(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event, mutatedData []byte) (fErr error) {
	handlerTimer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		status := metricspkg.StatusLabelSuccess
		if fErr != nil {
			status = metricspkg.StatusLabelError
		}
		handlerDuration.WithLabelValues(status, ref.ResourceID()).Observe(v)
	}))
	defer handlerTimer.ObserveDuration()

	handler, err := a.getHandlerAdapterForResource(ctx, ref)
	if err != nil {
		return err
	}

	return handler.Handle(ctx, ref, event, mutatedData)
}

func (a *AdapterV1) getHandlerAdapterForResource(ctx context.Context, ref *corev2.ResourceReference) (HandlerAdapter, error) {
	for _, handlerAdapter := range a.HandlerAdapters {
		if handlerAdapter.CanHandle(ref) {
			return handlerAdapter, nil
		}
	}
	return nil, fmt.Errorf("no handler adapters were found that can handle the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}
