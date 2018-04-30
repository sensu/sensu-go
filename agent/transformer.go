package agent

import "github.com/sensu/sensu-go/types"

// A Transformer handles transforming Sensu metrics to other metric formats
type Transformer interface {
	// Transform transforms a metric in a different metric format to Sensu Metric
	// Format
	Transform() []types.MetricPoint
}
