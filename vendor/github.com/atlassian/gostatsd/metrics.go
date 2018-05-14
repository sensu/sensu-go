package gostatsd

import (
	"bytes"
	"fmt"
	"hash/adler32"
)

// MetricType is an enumeration of all the possible types of Metric.
type MetricType byte

const (
	_ = iota
	// COUNTER is statsd counter type
	COUNTER MetricType = iota
	// TIMER is statsd timer type
	TIMER
	// GAUGE is statsd gauge type
	GAUGE
	// SET is statsd set type
	SET
)

func (m MetricType) String() string {
	switch m {
	case SET:
		return "set"
	case GAUGE:
		return "gauge"
	case TIMER:
		return "timer"
	case COUNTER:
		return "counter"
	}
	return "unknown"
}

// Metric represents a single data collected datapoint.
type Metric struct {
	Name        string     // The name of the metric
	Value       float64    // The numeric value of the metric
	Tags        Tags       // The tags for the metric
	TagsKey     string     // The tags rendered as a string to uniquely identify the tagset in a map
	StringValue string     // The string value for some metrics e.g. Set
	Hostname    string     // Hostname of the source of the metric
	SourceIP    IP         // IP of the source of the metric
	Type        MetricType // The type of metric
	DoneFunc    func()     // Returns the metric to the pool. May be nil. Call Metric.Done(), not this.
}

// Reset is used to reset a metric to as clean state, called on re-use from the pool.
func (m *Metric) Reset() {
	m.Name = ""
	m.Value = 0
	m.Tags = m.Tags[:0]
	m.TagsKey = ""
	m.StringValue = ""
	m.Hostname = ""
	m.SourceIP = ""
	m.Type = 0
}

// Bucket will pick a distribution bucket for this metric to land in.  max is exclusive.
func (m *Metric) Bucket(max int) int {
	bucket := adler32.Checksum([]byte(m.Name))
	bucket += adler32.Checksum([]byte(m.Hostname))
	// Consider hashing the tags here too
	bucket %= uint32(max)
	return int(bucket)
}

func (m *Metric) String() string {
	return fmt.Sprintf("{%s, %s, %f, %s, %v}", m.Type, m.Name, m.Value, m.StringValue, m.Tags)
}

// Done invokes DoneFunc if it's set, returning the metric to the pool.
func (m *Metric) Done() {
	if m.DoneFunc != nil {
		m.DoneFunc()
	}
}

// AggregatedMetrics is an interface for aggregated metrics.
type AggregatedMetrics interface {
	MetricsName() string
	Delete(string)
	DeleteChild(string, string)
	HasChildren(string) bool
}

// MetricMap is used for storing aggregated Metric values.
// The keys of each map are metric names.
type MetricMap struct {
	Counters Counters
	Timers   Timers
	Gauges   Gauges
	Sets     Sets
}

func (m *MetricMap) String() string {
	buf := new(bytes.Buffer)
	m.Counters.Each(func(k, tags string, counter Counter) {
		fmt.Fprintf(buf, "stats.counter.%s: %d tags=%s\n", k, counter.Value, tags)
	})
	m.Timers.Each(func(k, tags string, timer Timer) {
		for _, value := range timer.Values {
			fmt.Fprintf(buf, "stats.timer.%s: %f tags=%s\n", k, value, tags)
		}
	})
	m.Gauges.Each(func(k, tags string, gauge Gauge) {
		fmt.Fprintf(buf, "stats.gauge.%s: %f tags=%s\n", k, gauge.Value, tags)
	})
	m.Sets.Each(func(k, tags string, set Set) {
		fmt.Fprintf(buf, "stats.set.%s: %d tags=%s\n", k, len(set.Values), tags)
	})
	return buf.String()
}
