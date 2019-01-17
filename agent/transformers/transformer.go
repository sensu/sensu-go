package transformers

import "errors"

// ErrMetricExtraction is a blanket error for when transformer fails to extract metrics from a single line protocol
var ErrMetricExtraction = errors.New("unable to extract metric from check output")

// Field is a key value pair representing a metric
type Field struct {
	Key   string
	Value float64
}
