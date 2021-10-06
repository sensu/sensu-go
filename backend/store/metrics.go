package store

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sensu/sensu-go/backend/metrics"
	"github.com/sirupsen/logrus"
)

const (

	// EventBytesSummaryName is the name of the prometheus summary vec used to
	// track event sizes (in bytes).
	EventBytesSummaryName = "sensu_go_store_event_bytes"

	// EventBytesSummaryHelp is the help message for EventBytesSummary
	// Prometheus metrics.
	EventBytesSummaryHelp = "Distribution of event sizes, in bytes, received by the store on this backend"
)

var (
	EventBytesSummary = metrics.NewEventBytesSummaryVec(EventBytesSummaryName, EventBytesSummaryHelp)
	logger            = logrus.WithFields(logrus.Fields{
		"component": "store",
	})
)

func init() {
	if err := prometheus.Register(EventBytesSummary); err != nil {
		metrics.LogError(logger, EventBytesSummaryName, err)
	}
}
