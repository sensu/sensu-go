package datadog

import (
	"fmt"

	"github.com/atlassian/gostatsd"
)

type metricType string

const (
	// gauge is datadog gauge type.
	gauge metricType = "gauge"
	// rate is datadog rate type.
	rate metricType = "rate"
)

// flush represents a send operation.
type flush struct {
	ts               *timeSeries
	timestamp        float64
	flushIntervalSec float64
	metricsPerBatch  uint
	cb               func(*timeSeries)
}

// timeSeries represents a time series data structure.
type timeSeries struct {
	Series []metric `json:"series"`
}

// metric represents a metric data structure for Datadog.
type metric struct {
	Host     string     `json:"host,omitempty"`
	Interval float64    `json:"interval,omitempty"`
	Metric   string     `json:"metric"`
	Points   [1]point   `json:"points"`
	Tags     []string   `json:"tags,omitempty"`
	Type     metricType `json:"type,omitempty"`
}

// point is a Datadog data point.
type point [2]float64

// addMetricf adds a metric to the series.
func (f *flush) addMetricf(metricType metricType, value float64, hostname string, tags gostatsd.Tags, nameFormat string, a ...interface{}) {
	f.addMetric(metricType, value, hostname, tags, fmt.Sprintf(nameFormat, a...))
}

// addMetric adds a metric to the series.
func (f *flush) addMetric(metricType metricType, value float64, hostname string, tags gostatsd.Tags, name string) {
	f.ts.Series = append(f.ts.Series, metric{
		Host:     hostname,
		Interval: f.flushIntervalSec,
		Metric:   name,
		Points:   [1]point{{f.timestamp, value}},
		Tags:     tags,
		Type:     metricType,
	})
}

func (f *flush) maybeFlush() {
	if uint(len(f.ts.Series))+20 >= f.metricsPerBatch { // flush before it reaches max size and grows the slice
		f.cb(f.ts)
		f.ts = &timeSeries{
			Series: make([]metric, 0, f.metricsPerBatch),
		}
	}
}

func (f *flush) finish() {
	if len(f.ts.Series) > 0 {
		f.cb(f.ts)
	}
}
