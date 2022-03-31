package v2

import (
	math "math"
	"time"
)

// Validate returns an error if metrics does not pass validation tests.
func (m *Metrics) Validate() error {
	return nil
}

// TimestampAsNanoseconds will return the timestamp of the MetricPoint with
// nanosecond precision.
func (m *MetricPoint) TimestampAsNanoseconds() int64 {
	timestamp := m.Timestamp
	switch ts := math.Log10(float64(timestamp)); {
	case ts < 10:
		// assume timestamp is seconds
		timestamp = time.Unix(timestamp, 0).UnixNano() / int64(time.Nanosecond)
	case ts < 13:
		// assume timestamp is milliseconds
	case ts < 16:
		// assume timestamp is microseconds
		timestamp = (timestamp * 1000) / int64(time.Nanosecond)
	}

	return timestamp
}

func (m *MetricPoint) TimestampAsMilliseconds() int64 {
	timestamp := ts
	switch ts := math.Log10(float64(timestamp)); {
	case ts < 10:
		// assume timestamp is seconds
		timestamp = time.Unix(timestamp, 0).UnixNano() / int64(time.Millisecond)
	case ts < 13:
		// assume timestamp is milliseconds
	case ts < 16:
		// assume timestamp is microseconds
		timestamp = (timestamp * 1000) / int64(time.Millisecond)
	default:
		// assume timestamp is nanoseconds
		timestamp = timestamp / int64(time.Millisecond)
	}

	return timestamp
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
