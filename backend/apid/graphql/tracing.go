package graphql

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sensu/sensu-go/graphql/tracing"
)

func init() {
	if err := prometheus.Register(tracing.Collector); err != nil {
		logger.WithError(err).Error("unable to register tracer")
	}
}
