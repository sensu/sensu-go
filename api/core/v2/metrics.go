package v2

import (
	"time"
)

// Validate returns an error if metrics does not pass validation tests.
func (m *Metrics) Validate() error {
	return nil
}

// FixtureMetrics returns a testing fixture for a Metrics object.
func FixtureMetrics() *Metrics {
	return &Metrics{
		Handlers: []string{"influxdb"},
		Points:   []*MetricPoint{FixtureMetricPoint()},
	}
}

// FixtureMetricPoint returns a testing fixture for a Metric Point object.
func FixtureMetricPoint() *MetricPoint {
	return &MetricPoint{
		Name:      "answer",
		Value:     42.0,
		Timestamp: time.Now().UnixNano(),
		Tags:      []*MetricTag{FixtureMetricTag()},
	}
}

// FixtureMetricTag returns a testing fixture for a Metric Tag object.
func FixtureMetricTag() *MetricTag {
	return &MetricTag{
		Name:  "foo",
		Value: "bar",
	}
}
