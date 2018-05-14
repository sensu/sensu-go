// +build integration

package agent

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/atlassian/gostatsd"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	counter := FixtureCounter(now)

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
	timer := FixtureTimer(now)

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
	gauge := FixtureGauge(now)

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
	set := FixtureSet(now)

	points := composeSetPoints(set, key, tags, now)
	assert.Equal(t, len(points), 1)
	for _, point := range points {
		assert.Contains(t, point.Name, key)
		assert.NotNil(t, point.Value)
		assert.Equal(t, point.Timestamp, now)
		assert.Equal(t, point.Tags, tags)
	}
}

func TestPreparePoints(t *testing.T) {
	now := time.Now().UnixNano()
	metrics := FixtureMetricMap(now)

	points := prepareMetrics(now, &metrics)
	assert.Equal(t, len(points), 14)
	for _, point := range points {
		assert.Contains(t, point.Name, "test")
		assert.Equal(t, point.Timestamp, now)
	}
}

func TestReceiveMetrics(t *testing.T) {
	assert := assert.New(t)

	cfg := FixtureConfig()
	cfg.StatsdServer.FlushInterval = 1
	ta := NewAgent(cfg)

	go ta.StartStatsd()
	// Give the server a second to start up
	time.Sleep(time.Second * 1)

	udpClient, err := net.Dial("udp", ta.statsdServer.MetricsAddr)
	if err != nil {
		assert.FailNow("failed to create UDP connection")
	}

	_, err = udpClient.Write([]byte("foo:1|c"))
	require.NoError(t, err)
	require.NoError(t, udpClient.Close())

	msg := <-ta.sendq
	assert.NotEmpty(msg)
	assert.Equal("event", msg.Type)

	var event types.Event
	err = json.Unmarshal(msg.Payload, &event)
	if err != nil {
		assert.FailNow("failed to unmarshal event json")
	}

	assert.NotNil(event.Entity)
	assert.NotNil(event.Metrics)
	assert.Nil(event.Check)
	for _, point := range event.Metrics.Points {
		assert.Contains(point.Name, "foo")
	}
}

func FixtureCounter(now int64) gostatsd.Counter {
	return gostatsd.Counter{
		PerSecond: 2,
		Value:     3,
		Timestamp: gostatsd.Nanotime(now),
		Hostname:  "host",
		Tags:      gostatsd.Tags{"foo:bar"},
	}
}

func FixtureCounters(now int64) gostatsd.Counters {
	counters := make(map[string]map[string]gostatsd.Counter)
	counter := make(map[string]gostatsd.Counter)
	c := FixtureCounter(now)
	counter["c1"] = c
	counters["test"] = counter
	return counters
}

func FixtureTimer(now int64) gostatsd.Timer {
	return gostatsd.Timer{
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
}

func FixtureTimers(now int64) gostatsd.Timers {
	timers := make(map[string]map[string]gostatsd.Timer)
	timer := make(map[string]gostatsd.Timer)
	t := FixtureTimer(now)
	timer["t1"] = t
	timers["test"] = timer
	return timers
}

func FixtureGauge(now int64) gostatsd.Gauge {
	return gostatsd.Gauge{
		Value:     3,
		Timestamp: gostatsd.Nanotime(now),
		Hostname:  "host",
		Tags:      gostatsd.Tags{"foo:bar"},
	}
}

func FixtureGauges(now int64) gostatsd.Gauges {
	gauges := make(map[string]map[string]gostatsd.Gauge)
	gauge := make(map[string]gostatsd.Gauge)
	g := FixtureGauge(now)
	gauge["g1"] = g
	gauges["test"] = gauge
	return gauges
}

func FixtureSet(now int64) gostatsd.Set {
	return gostatsd.Set{
		Values:    map[string]struct{}{"foo": struct{}{}},
		Timestamp: gostatsd.Nanotime(now),
		Hostname:  "host",
		Tags:      gostatsd.Tags{"foo:bar"},
	}
}

func FixtureSets(now int64) gostatsd.Sets {
	sets := make(map[string]map[string]gostatsd.Set)
	set := make(map[string]gostatsd.Set)
	s := FixtureSet(now)
	set["s1"] = s
	sets["test"] = set
	return sets
}

func FixtureMetricMap(now int64) gostatsd.MetricMap {
	return gostatsd.MetricMap{
		Counters: FixtureCounters(now),
		Timers:   FixtureTimers(now),
		Gauges:   FixtureGauges(now),
		Sets:     FixtureSets(now),
	}
}
