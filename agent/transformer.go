package agent

import corev2 "github.com/sensu/core/v2"

// A Transformer handles transforming Sensu metrics to other output metric formats
type Transformer interface {
	// Transform transforms a metric in a different output metric format to Sensu Metric
	// Format
	Transform() []*corev2.MetricPoint
}
