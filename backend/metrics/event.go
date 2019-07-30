package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	// EventsProcessedCounterVec is the name of the prometheus counter vec used to count events processed.
	EventsProcessedCounterVec = "sensu_go_events_processed"

	// EventsProcessedLabelName is the name of the label which stores prometheus values.
	EventsProcessedLabelName = "status"

	// EventsProcessedLabelSuccess is the name of the label used to count events processed successfully.
	EventsProcessedLabelSuccess = "success"
)

var (
	// EventsProcessed counts the number of sensu go events processed.
	EventsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: EventsProcessedCounterVec,
			Help: "The total number of processed events",
		},
		[]string{EventsProcessedLabelName},
	)
)

func init() {
	_ = prometheus.Register(EventsProcessed)
}
