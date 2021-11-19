package pipeline

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	metricspkg "github.com/sensu/sensu-go/metrics"
)

const (
	// MutatorDuration is the name of the prometheus summary vec used to track
	// average latencies of pipeline mutator execution.
	MutatorDuration = "sensu_go_pipeline_mutator_duration"
)

var (
	mutatorDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       MutatorDuration,
			Help:       "pipeline mutator execution latency distribution",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{metricspkg.StatusLabelName, metricspkg.ResourceReferenceLabelName},
	)
)

type MutatorAdapter interface {
	Name() string
	CanMutate(*corev2.ResourceReference) bool
	Mutate(context.Context, *corev2.ResourceReference, *corev2.Event) ([]byte, error)
}

func init() {
	if err := prometheus.Register(mutatorDuration); err != nil {
		panic(fmt.Errorf("error registering %s: %s", MutatorDuration, err))
	}
}

func (a *AdapterV1) processMutator(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (data []byte, fErr error) {
	mutatorTimer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		status := metricspkg.StatusLabelSuccess
		if fErr != nil {
			status = metricspkg.StatusLabelError
		}
		mutatorDuration.WithLabelValues(status, ref.ResourceID()).Observe(v * float64(1000))
	}))
	defer mutatorTimer.ObserveDuration()

	mutator, err := a.getMutatorAdapterForResource(ctx, ref)
	if err != nil {
		return nil, err
	}

	return mutator.Mutate(ctx, ref, event)
}

func (a *AdapterV1) getMutatorAdapterForResource(ctx context.Context, ref *corev2.ResourceReference) (MutatorAdapter, error) {
	for _, mutatorAdapter := range a.MutatorAdapters {
		if mutatorAdapter.CanMutate(ref) {
			return mutatorAdapter, nil
		}
	}
	return nil, fmt.Errorf("no mutator adapters were found that can mutate the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}
