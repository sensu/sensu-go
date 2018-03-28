package agent

import (
	"testing"
	"time"

	"github.com/atlassian/gostatsd"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestNewStatsdServer(t *testing.T) {
	c := FixtureConfig()
	a := &Agent{config: c}
	s := NewStatsdServer(a)
	assert.NotNil(t, s)
	assert.Equal(t, BackendName, s.Backends[0].Name())
	assert.Equal(t, DefaultStatsdFlushInterval*time.Second, s.FlushInterval)
	assert.Equal(t, "127.0.0.1:8125", s.MetricsAddr)

	c.StatsdServer.FlushInterval = 20
	c.StatsdServer.Port = 8126
	c.StatsdServer.Host = "foo"
	a.config = c
	s = NewStatsdServer(a)
	assert.NotNil(t, s)
	assert.Equal(t, BackendName, s.Backends[0].Name())
	assert.Equal(t, 20*time.Second, s.FlushInterval)
	assert.Equal(t, "foo:8126", s.MetricsAddr)
}

func TestComposeMetricTags(t *testing.T) {
	testCases := []struct {
		name      string
		tagsKey   string
		metricTag []*types.MetricTag
	}{
		{
			name:    "Full tagsKey",
			tagsKey: "aggregator_id:5,channel:dispatch_aggregator,s:hostName",
			metricTag: []*types.MetricTag{
				{Name: "aggregator_id", Value: "5"},
				{Name: "channel", Value: "dispatch_aggregator"},
				{Name: "s", Value: "hostName"},
			},
		},
		{
			name:    "Single tagsKey",
			tagsKey: "aggregator_id:5",
			metricTag: []*types.MetricTag{
				{Name: "aggregator_id", Value: "5"},
			},
		},
		{
			name:      "Empty tagsKey",
			tagsKey:   "",
			metricTag: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metricTag := composeMetricTags(tc.tagsKey)
			assert.Equal(t, tc.metricTag, metricTag)
		})
	}
}

func TestComposeCounterPoints(t *testing.T) {
	now := time.Now().UnixNano()
	key := "foo:bar"
	tags := composeMetricTags("boo:baz")
	counter := gostatsd.Counter{
		PerSecond: 2,
		Value:     3,
		Timestamp: gostatsd.Nanotime(now),
		Hostname:  "host",
		Tags:      gostatsd.Tags{"foo:bar"},
	}

	points := composeCounterPoints(counter, key, tags, now)
	assert.Equal(t, len(points), 2)
	for _, point := range points {
		assert.Contains(t, point.Name, key)
		assert.NotNil(t, point.Value)
		assert.Equal(t, point.Timestamp, now)
		assert.Equal(t, point.Tags, tags)
	}
}

func TestComposeTimerPoints(t *testing.T) {
	now := time.Now().UnixNano()
	key := "foo:bar"
	tags := composeMetricTags("boo:baz")
	timer := gostatsd.Timer{
		Count:       2,
		PerSecond:   3,
		Mean:        4,
		Median:      5,
		Min:         6,
		Max:         7,
		StdDev:      8,
		Sum:         9,
		SumSquares:  10,
		Values:      []float64{1, 2, 3},
		Percentiles: []gostatsd.Percentile{{Float: 4, Str: "str"}},
		Timestamp:   gostatsd.Nanotime(now),
		Hostname:    "host",
		Tags:        gostatsd.Tags{"foo:bar"},
	}

	points := composeTimerPoints(timer, key, tags, now)
	assert.Equal(t, len(points), 10)
	for _, point := range points {
		assert.Contains(t, point.Name, key)
		assert.NotNil(t, point.Value)
		assert.Equal(t, point.Timestamp, now)
		assert.Equal(t, point.Tags, tags)
	}
}

func TestComposeGaugePoints(t *testing.T) {
	now := time.Now().UnixNano()
	key := "foo:bar"
	tags := composeMetricTags("boo:baz")
	gauge := gostatsd.Gauge{
		Value:     3,
		Timestamp: gostatsd.Nanotime(now),
		Hostname:  "host",
		Tags:      gostatsd.Tags{"foo:bar"},
	}

	points := composeGaugePoints(gauge, key, tags, now)
	assert.Equal(t, len(points), 1)
	for _, point := range points {
		assert.Contains(t, point.Name, key)
		assert.NotNil(t, point.Value)
		assert.Equal(t, point.Timestamp, now)
		assert.Equal(t, point.Tags, tags)
	}
}

func TestComposeSetPoints(t *testing.T) {
	now := time.Now().UnixNano()
	key := "foo:bar"
	tags := composeMetricTags("boo:baz")
	set := gostatsd.Set{
		Values:    map[string]struct{}{"foo": struct{}{}},
		Timestamp: gostatsd.Nanotime(now),
		Hostname:  "host",
		Tags:      gostatsd.Tags{"foo:bar"},
	}

	points := composeSetPoints(set, key, tags, now)
	assert.Equal(t, len(points), 1)
	for _, point := range points {
		assert.Contains(t, point.Name, key)
		assert.NotNil(t, point.Value)
		assert.Equal(t, point.Timestamp, now)
		assert.Equal(t, point.Tags, tags)
	}
}
