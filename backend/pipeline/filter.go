package pipeline

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	corev2 "github.com/sensu/core/v2"
	metricspkg "github.com/sensu/sensu-go/metrics"
)

const (
	// FilterDuration is the name of the prometheus summary vec used to track
	// average latencies of pipeline filter execution.
	FilterDuration = "sensu_go_pipeline_filter_duration"
)

var (
	filterDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       FilterDuration,
			Help:       "pipeline filter execution latency distribution",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{metricspkg.StatusLabelName, metricspkg.ResourceReferenceLabelName},
	)
)

type FilterAdapter interface {
	Name() string
	CanFilter(*corev2.ResourceReference) bool
	Filter(context.Context, *corev2.ResourceReference, *corev2.Event) (bool, error)
}

func init() {
	if err := prometheus.Register(filterDuration); err != nil {
		panic(fmt.Errorf("error registering %s: %s", FilterDuration, err))
	}
}

func (a *AdapterV1) processFilters(ctx context.Context, refs []*corev2.ResourceReference, event *corev2.Event) (bool, error) {
	// for each filter reference in the workflow, attempt to find a compatible
	// filter adapter and use it to filter the event.
	for _, ref := range refs {
		filtered, err := a.processFilter(ctx, ref, event)
		if err != nil {
			return false, err
		}
		if filtered {
			return true, nil
		}
	}
	return false, nil
}

func (a *AdapterV1) processFilter(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (filtered bool, fErr error) {
	filterTimer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		status := metricspkg.StatusLabelSuccess
		if fErr != nil {
			status = metricspkg.StatusLabelError
		}
		filterDuration.WithLabelValues(status, ref.ResourceID()).Observe(v * float64(1000))
	}))
	defer filterTimer.ObserveDuration()

	filter, err := a.getFilterAdapterForResource(ctx, ref)
	if err != nil {
		return false, err
	}

	return filter.Filter(ctx, ref, event)
}

func (a *AdapterV1) getFilterAdapterForResource(ctx context.Context, ref *corev2.ResourceReference) (FilterAdapter, error) {
	for _, filterAdapter := range a.FilterAdapters {
		if filterAdapter.CanFilter(ref) {
			return filterAdapter, nil
		}
	}
	return nil, fmt.Errorf("no filter adapters were found that can filter the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}
