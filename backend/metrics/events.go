package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	// EventTypeLabelName is the name of the label for event type.
	EventTypeLabelName = "type"

	// EventTypeLabelCheck is the event type name for events with only checks.
	EventTypeLabelCheck = "check"

	// EventTypeLabelMetrics is the event type name for events with only
	// metrics.
	EventTypeLabelMetrics = "metrics"

	// EventTypeLabelCheckAndMetrics is the event type name for events with a
	// check & metrics.
	EventTypeLabelCheckAndMetrics = "check+metrics"
)

func NewEventBytesSummaryVec(name, help string) *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: name,
			Help: help,
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{EventTypeLabelName},
	)
}
