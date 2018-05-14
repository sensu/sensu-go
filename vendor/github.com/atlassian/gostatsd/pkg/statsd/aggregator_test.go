package statsd

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/ash2k/stager"
	"github.com/stretchr/testify/assert"

	"github.com/atlassian/gostatsd"
)

func newFakeAggregator() *MetricAggregator {
	return NewMetricAggregator(
		[]float64{90},
		5*time.Minute,
		gostatsd.TimerSubtypes{},
	)
}

type fakeAggregatorFactory struct{}

func (faf *fakeAggregatorFactory) Create() Aggregator {
	return newFakeAggregator()
}

func TestNewAggregator(t *testing.T) {
	t.Parallel()
	assrt := assert.New(t)

	actual := newFakeAggregator()

	if assrt.NotNil(actual.Counters) {
		assrt.Equal(gostatsd.Counters{}, actual.Counters)
	}

	if assrt.NotNil(actual.Timers) {
		assrt.Equal(gostatsd.Timers{}, actual.Timers)
	}

	if assrt.NotNil(actual.Gauges) {
		assrt.Equal(gostatsd.Gauges{}, actual.Gauges)
	}

	if assrt.NotNil(actual.Sets) {
		assrt.Equal(gostatsd.Sets{}, actual.Sets)
	}
}

func TestFlush(t *testing.T) {
	t.Parallel()
	assrt := assert.New(t)

	now := time.Now()
	nowFn := func() time.Time { return now }
	ma := newFakeAggregator()
	ma.now = nowFn
	expected := newFakeAggregator()
	expected.now = nowFn

	ma.Counters["some"] = make(map[string]gostatsd.Counter)
	ma.Counters["some"][""] = gostatsd.Counter{Value: 50}
	ma.Counters["some"]["thing"] = gostatsd.Counter{Value: 100}
	ma.Counters["some"]["other:thing"] = gostatsd.Counter{Value: 150}

	expected.Counters["some"] = make(map[string]gostatsd.Counter)
	expected.Counters["some"][""] = gostatsd.Counter{Value: 50, PerSecond: 5}
	expected.Counters["some"]["thing"] = gostatsd.Counter{Value: 100, PerSecond: 10}
	expected.Counters["some"]["other:thing"] = gostatsd.Counter{Value: 150, PerSecond: 15}

	ma.Timers["some"] = make(map[string]gostatsd.Timer)
	ma.Timers["some"]["thing"] = gostatsd.Timer{Values: []float64{2, 4, 12}}
	ma.Timers["some"]["empty"] = gostatsd.Timer{Values: []float64{}}

	expPct := gostatsd.Percentiles{}
	expPct.Set("count_90", float64(3))
	expPct.Set("mean_90", float64(6))
	expPct.Set("sum_90", float64(18))
	expPct.Set("sum_squares_90", float64(164))
	expPct.Set("upper_90", float64(12))
	expected.Timers["some"] = make(map[string]gostatsd.Timer)
	expected.Timers["some"]["thing"] = gostatsd.Timer{
		Values: []float64{2, 4, 12}, Count: 3, Min: 2, Max: 12, Mean: 6, Median: 4, Sum: 18,
		PerSecond: 0.3, SumSquares: 164, StdDev: 4.320493798938574, Percentiles: expPct,
	}
	expected.Timers["some"]["empty"] = gostatsd.Timer{Values: []float64{}}

	ma.Gauges["some"] = make(map[string]gostatsd.Gauge)
	ma.Gauges["some"][""] = gostatsd.Gauge{Value: 50}
	ma.Gauges["some"]["thing"] = gostatsd.Gauge{Value: 100}
	ma.Gauges["some"]["other:thing"] = gostatsd.Gauge{Value: 150}

	expected.Gauges["some"] = make(map[string]gostatsd.Gauge)
	expected.Gauges["some"][""] = gostatsd.Gauge{Value: 50}
	expected.Gauges["some"]["thing"] = gostatsd.Gauge{Value: 100}
	expected.Gauges["some"]["other:thing"] = gostatsd.Gauge{Value: 150}

	ma.Sets["some"] = make(map[string]gostatsd.Set)
	unique := map[string]struct{}{
		"user": {},
	}
	ma.Sets["some"]["thing"] = gostatsd.Set{Values: unique}

	expected.Sets["some"] = make(map[string]gostatsd.Set)
	expected.Sets["some"]["thing"] = gostatsd.Set{Values: unique}

	ma.Flush(10 * time.Second)
	assrt.Equal(expected.Counters, ma.Counters)
	assrt.Equal(expected.Timers, ma.Timers)
	assrt.Equal(expected.Gauges, ma.Gauges)
	assrt.Equal(expected.Sets, ma.Sets)
}

func BenchmarkFlush(b *testing.B) {
	ma := newFakeAggregator()
	ma.Counters["some"] = make(map[string]gostatsd.Counter)
	ma.Counters["some"][""] = gostatsd.Counter{Value: 50}
	ma.Counters["some"]["thing"] = gostatsd.Counter{Value: 100}
	ma.Counters["some"]["other:thing"] = gostatsd.Counter{Value: 150}

	ma.Timers["some"] = make(map[string]gostatsd.Timer)
	ma.Timers["some"]["thing"] = gostatsd.Timer{Values: []float64{2, 4, 12}}
	ma.Timers["some"]["empty"] = gostatsd.Timer{Values: []float64{}}

	ma.Gauges["some"] = make(map[string]gostatsd.Gauge)
	ma.Gauges["some"][""] = gostatsd.Gauge{Value: 50}
	ma.Gauges["some"]["thing"] = gostatsd.Gauge{Value: 100}
	ma.Gauges["some"]["other:thing"] = gostatsd.Gauge{Value: 150}

	ma.Sets["some"] = make(map[string]gostatsd.Set)
	unique := map[string]struct{}{
		"user": {},
	}
	ma.Sets["some"]["thing"] = gostatsd.Set{Values: unique}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		ma.Flush(1 * time.Second)
	}
}

func TestReset(t *testing.T) {
	t.Parallel()
	assrt := assert.New(t)
	now := time.Now()
	nowNano := gostatsd.Nanotime(now.UnixNano())
	nowFn := func() time.Time { return now }
	host := "hostname"

	// non expired
	actual := newFakeAggregator()
	actual.Counters["some"] = map[string]gostatsd.Counter{
		"thing":       gostatsd.NewCounter(nowNano, 50, host, nil),
		"other:thing": gostatsd.NewCounter(nowNano, 90, host, nil),
	}
	actual.now = nowFn
	actual.Reset()

	expected := newFakeAggregator()
	expected.Counters["some"] = map[string]gostatsd.Counter{
		"thing":       gostatsd.NewCounter(nowNano, 0, host, nil),
		"other:thing": gostatsd.NewCounter(nowNano, 0, host, nil),
	}
	expected.now = nowFn

	assrt.Equal(expected.Counters, actual.Counters)

	actual = newFakeAggregator()
	actual.Timers["some"] = map[string]gostatsd.Timer{
		"thing": gostatsd.NewTimer(nowNano, []float64{50}, host, nil),
	}
	actual.now = nowFn
	actual.Reset()

	expected = newFakeAggregator()
	expected.Timers["some"] = map[string]gostatsd.Timer{
		"thing": gostatsd.NewTimer(nowNano, []float64{}, host, nil),
	}
	expected.now = nowFn

	assrt.Equal(expected.Timers, actual.Timers)

	actual = newFakeAggregator()
	actual.Gauges["some"] = map[string]gostatsd.Gauge{
		"thing":       gostatsd.NewGauge(nowNano, 50, host, nil),
		"other:thing": gostatsd.NewGauge(nowNano, 90, host, nil),
	}
	actual.now = nowFn
	actual.Reset()

	expected = newFakeAggregator()
	expected.Gauges["some"] = map[string]gostatsd.Gauge{
		"thing":       gostatsd.NewGauge(nowNano, 50, host, nil),
		"other:thing": gostatsd.NewGauge(nowNano, 90, host, nil),
	}
	expected.now = nowFn

	assrt.Equal(expected.Gauges, actual.Gauges)

	actual = newFakeAggregator()
	actual.Sets["some"] = map[string]gostatsd.Set{
		"thing": gostatsd.NewSet(nowNano, map[string]struct{}{"user": {}}, host, nil),
	}
	actual.now = nowFn
	actual.Reset()

	expected = newFakeAggregator()
	expected.Sets["some"] = map[string]gostatsd.Set{
		"thing": gostatsd.NewSet(nowNano, make(map[string]struct{}), host, nil),
	}
	expected.now = nowFn

	assrt.Equal(expected.Sets, actual.Sets)

	// expired
	pastNano := gostatsd.Nanotime(now.Add(-30 * time.Second).UnixNano())

	actual = newFakeAggregator()
	actual.expiryInterval = 10 * time.Second
	actual.Counters["some"] = map[string]gostatsd.Counter{
		"thing":       gostatsd.NewCounter(pastNano, 50, host, nil),
		"other:thing": gostatsd.NewCounter(pastNano, 90, host, nil),
	}
	actual.now = nowFn
	actual.Reset()

	expected = newFakeAggregator()
	expected.now = nowFn

	assrt.Equal(expected.Counters, actual.Counters)

	actual = newFakeAggregator()
	actual.expiryInterval = 10 * time.Second
	actual.Timers["some"] = map[string]gostatsd.Timer{
		"thing": gostatsd.NewTimer(pastNano, []float64{50}, host, nil),
	}
	actual.now = nowFn
	actual.Reset()

	expected = newFakeAggregator()
	expected.now = nowFn

	assrt.Equal(expected.Timers, actual.Timers)

	actual = newFakeAggregator()
	actual.expiryInterval = 10 * time.Second
	actual.Gauges["some"] = map[string]gostatsd.Gauge{
		"thing":       gostatsd.NewGauge(pastNano, 50, host, nil),
		"other:thing": gostatsd.NewGauge(pastNano, 90, host, nil),
	}
	actual.now = nowFn
	actual.Reset()

	expected = newFakeAggregator()
	expected.now = nowFn

	assrt.Equal(expected.Gauges, actual.Gauges)

	actual = newFakeAggregator()
	actual.expiryInterval = 10 * time.Second
	actual.Sets["some"] = map[string]gostatsd.Set{
		"thing": gostatsd.NewSet(pastNano, map[string]struct{}{"user": {}}, host, nil),
	}
	actual.now = nowFn
	actual.Reset()

	expected = newFakeAggregator()
	expected.now = nowFn

	assrt.Equal(expected.Sets, actual.Sets)
}

func TestIsExpired(t *testing.T) {
	t.Parallel()
	assrt := assert.New(t)

	now := gostatsd.Nanotime(time.Now().UnixNano())

	ma := &MetricAggregator{expiryInterval: 0}
	assrt.Equal(false, ma.isExpired(now, now))

	ma.expiryInterval = 10 * time.Second

	ts := gostatsd.Nanotime(time.Now().Add(-30 * time.Second).UnixNano())
	assrt.Equal(true, ma.isExpired(now, ts))

	ts = gostatsd.Nanotime(time.Now().Add(-1 * time.Second).UnixNano())
	assrt.Equal(false, ma.isExpired(now, ts))
}

func TestDisabledCount(t *testing.T) {
	t.Parallel()
	ma := newFakeAggregator()
	ma.disabledSubtypes.CountPct = true
	ma.Receive(&gostatsd.Metric{Name: "x", Value: 1, Type: gostatsd.TIMER}, time.Now())
	ma.Flush(1 * time.Second)
	for _, pct := range ma.Timers["x"][""].Percentiles {
		if pct.Str == "count_90" {
			t.Error("count not disabled")
		}
	}
}

func TestDisabledMean(t *testing.T) {
	t.Parallel()
	ma := newFakeAggregator()
	ma.disabledSubtypes.MeanPct = true
	ma.Receive(&gostatsd.Metric{Name: "x", Value: 1, Type: gostatsd.TIMER}, time.Now())
	ma.Flush(1 * time.Second)
	for _, pct := range ma.Timers["x"][""].Percentiles {
		if pct.Str == "mean_90" {
			t.Error("mean not disabled")
		}
	}
}

func TestDisabledSum(t *testing.T) {
	t.Parallel()
	ma := newFakeAggregator()
	ma.disabledSubtypes.SumPct = true
	ma.Receive(&gostatsd.Metric{Name: "x", Value: 1, Type: gostatsd.TIMER}, time.Now())
	ma.Flush(1 * time.Second)
	for _, pct := range ma.Timers["x"][""].Percentiles {
		if pct.Str == "sum_90" {
			t.Error("sum not disabled")
		}
	}
}

func TestDisabledSumSquares(t *testing.T) {
	t.Parallel()
	ma := newFakeAggregator()
	ma.disabledSubtypes.SumSquaresPct = true
	ma.Receive(&gostatsd.Metric{Name: "x", Value: 1, Type: gostatsd.TIMER}, time.Now())
	ma.Flush(1 * time.Second)
	for _, pct := range ma.Timers["x"][""].Percentiles {
		if pct.Str == "sum_squares_90" {
			t.Error("sum_squares not disabled")
		}
	}
}

func TestDisabledUpper(t *testing.T) {
	t.Parallel()
	ma := newFakeAggregator()
	ma.disabledSubtypes.UpperPct = true
	ma.Receive(&gostatsd.Metric{Name: "x", Value: 1, Type: gostatsd.TIMER}, time.Now())
	ma.Flush(1 * time.Second)
	for _, pct := range ma.Timers["x"][""].Percentiles {
		if pct.Str == "upper_90" {
			t.Error("upper not disabled")
		}
	}
}

func TestDisabledLower(t *testing.T) {
	t.Parallel()
	ma := NewMetricAggregator(
		[]float64{-90},
		5*time.Minute,
		gostatsd.TimerSubtypes{},
	)
	ma.disabledSubtypes.LowerPct = true
	ma.Receive(&gostatsd.Metric{Name: "x", Value: 1, Type: gostatsd.TIMER}, time.Now())
	ma.Flush(1 * time.Second)
	for _, pct := range ma.Timers["x"][""].Percentiles {
		if pct.Str == "lower_-90" { // lower_-90?
			t.Error("lower not disabled")
		}
	}
}

func metricsFixtures() []gostatsd.Metric {
	ms := []gostatsd.Metric{
		{Name: "foo.bar.baz", Value: 2, Type: gostatsd.COUNTER},
		{Name: "abc.def.g", Value: 3, Type: gostatsd.GAUGE},
		{Name: "abc.def.g", Value: 8, Type: gostatsd.GAUGE, Tags: gostatsd.Tags{"foo:bar", "baz"}},
		{Name: "def.g", Value: 10, Type: gostatsd.TIMER},
		{Name: "def.g", Value: 1, Type: gostatsd.TIMER, Tags: gostatsd.Tags{"foo:bar", "baz"}},
		{Name: "smp.rte", Value: 50, Type: gostatsd.COUNTER},
		{Name: "smp.rte", Value: 50, Type: gostatsd.COUNTER, Tags: gostatsd.Tags{"foo:bar", "baz"}},
		{Name: "smp.rte", Value: 5, Type: gostatsd.COUNTER, Tags: gostatsd.Tags{"foo:bar", "baz"}},
		{Name: "uniq.usr", StringValue: "joe", Type: gostatsd.SET},
		{Name: "uniq.usr", StringValue: "joe", Type: gostatsd.SET},
		{Name: "uniq.usr", StringValue: "bob", Type: gostatsd.SET},
		{Name: "uniq.usr", StringValue: "john", Type: gostatsd.SET},
		{Name: "uniq.usr", StringValue: "john", Type: gostatsd.SET, Tags: gostatsd.Tags{"foo:bar", "baz"}},
	}
	for i, m := range ms {
		ms[i].TagsKey = formatTagsKey(m.Tags, m.Hostname)
	}
	return ms
}

func TestReceive(t *testing.T) {
	t.Parallel()
	assrt := assert.New(t)

	ma := newFakeAggregator()
	now := time.Now()
	nowNano := gostatsd.Nanotime(now.UnixNano())

	tests := metricsFixtures()
	for _, metric := range tests {
		ma.Receive(&metric, now)
	}

	expectedCounters := gostatsd.Counters{
		"foo.bar.baz": map[string]gostatsd.Counter{
			"": {Value: 2, Timestamp: nowNano},
		},
		"smp.rte": map[string]gostatsd.Counter{
			"":            {Value: 50, Timestamp: nowNano},
			"baz,foo:bar": {Value: 55, Timestamp: nowNano, Tags: gostatsd.Tags{"baz", "foo:bar"}},
		},
	}
	assrt.Equal(expectedCounters, ma.Counters)

	expectedGauges := gostatsd.Gauges{
		"abc.def.g": map[string]gostatsd.Gauge{
			"":            {Value: 3, Timestamp: nowNano},
			"baz,foo:bar": {Value: 8, Timestamp: nowNano, Tags: gostatsd.Tags{"baz", "foo:bar"}},
		},
	}
	assrt.Equal(expectedGauges, ma.Gauges)

	expectedTimers := gostatsd.Timers{
		"def.g": map[string]gostatsd.Timer{
			"":            {Values: []float64{10}, Timestamp: nowNano},
			"baz,foo:bar": {Values: []float64{1}, Timestamp: nowNano, Tags: gostatsd.Tags{"baz", "foo:bar"}},
		},
	}
	assrt.Equal(expectedTimers, ma.Timers)

	expectedSets := gostatsd.Sets{
		"uniq.usr": map[string]gostatsd.Set{
			"": {
				Values: map[string]struct{}{
					"joe":  {},
					"bob":  {},
					"john": {},
				},
				Timestamp: nowNano,
			},
			"baz,foo:bar": {
				Values: map[string]struct{}{
					"john": {},
				},
				Timestamp: nowNano,
				Tags:      gostatsd.Tags{"baz", "foo:bar"},
			},
		},
	}
	assrt.Equal(expectedSets, ma.Sets)
}

func BenchmarkHotMetric(b *testing.B) {
	beh := NewBackendHandler(
		nil,
		1000,
		runtime.NumCPU(),
		10000,
		&fakeAggregatorFactory{},
	)

	stgr := stager.New()
	stage := stgr.NextStage()
	stage.StartWithContext(beh.Run)
	stage = stgr.NextStage()

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < runtime.NumCPU(); i++ {
		stage.Start(func() {
			for n := 0; n < b.N; n++ {
				m := &gostatsd.Metric{
					Name:     "metric.name",
					Value:    5,
					Tags:     gostatsd.Tags{"aaaa:aaaa", "aaab:aaab", "aaac:aaac", "aaad:aaad", "aaae:aaae", "aaaf:aaaf"},
					Hostname: "local",
					Type:     gostatsd.GAUGE,
				}
				beh.DispatchMetric(ctx, m)
			}
		})
	}

	stgr.Shutdown()
}

func benchmarkReceive(metric gostatsd.Metric, b *testing.B) {
	ma := newFakeAggregator()
	now := time.Now()
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		ma.Receive(&metric, now)
	}
}

func BenchmarkReceiveCounter(b *testing.B) {
	benchmarkReceive(gostatsd.Metric{Name: "foo.bar.baz", Value: 2, Type: gostatsd.COUNTER}, b)
}

func BenchmarkReceiveGauge(b *testing.B) {
	benchmarkReceive(gostatsd.Metric{Name: "abc.def.g", Value: 3, Type: gostatsd.GAUGE}, b)
}

func BenchmarkReceiveTimer(b *testing.B) {
	benchmarkReceive(gostatsd.Metric{Name: "def.g", Value: 10, Type: gostatsd.TIMER}, b)
}

func BenchmarkReceiveSet(b *testing.B) {
	benchmarkReceive(gostatsd.Metric{Name: "uniq.usr", StringValue: "joe", Type: gostatsd.SET}, b)
}

func BenchmarkReceives(b *testing.B) {
	ma := newFakeAggregator()
	now := time.Now()
	tests := metricsFixtures()
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, metric := range tests {
			ma.Receive(&metric, now)
		}
	}
}
