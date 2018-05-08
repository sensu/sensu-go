package agent

import "github.com/sensu/sensu-go/types"

// A Transformer handles transforming Sensu metrics to other output metric formats
type Transformer interface {
	// Transform transforms a metric in a different output metric format to Sensu Metric
	// Format
	Transform() []*types.MetricPoint
}
