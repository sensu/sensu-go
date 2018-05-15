package statsd

import (
	"context"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/atlassian/gostatsd"
	"github.com/atlassian/gostatsd/pkg/statser"

	log "github.com/sirupsen/logrus"
)

// percentStruct is a cache of percentile names to avoid creating them for each timer.
type percentStruct struct {
	count      string
	mean       string
	sum        string
	sumSquares string
	upper      string
	lower      string
}

// MetricAggregator aggregates metrics.
type MetricAggregator struct {
	metricsReceived   uint64
	expiryInterval    time.Duration // How often to expire metrics
	percentThresholds map[float64]percentStruct
	now               func() time.Time // Returns current time. Useful for testing.
	statser           statser.Statser
	disabledSubtypes  gostatsd.TimerSubtypes
	gostatsd.MetricMap
}

// NewMetricAggregator creates a new MetricAggregator object.
func NewMetricAggregator(percentThresholds []float64, expiryInterval time.Duration, disabled gostatsd.TimerSubtypes) *MetricAggregator {
	a := MetricAggregator{
		expiryInterval:    expiryInterval,
		percentThresholds: make(map[float64]percentStruct, len(percentThresholds)),
		now:               time.Now,
		statser:           statser.NewNullStatser(), // Will probably be replaced via RunMetrics
		MetricMap: gostatsd.MetricMap{
			Counters: gostatsd.Counters{},
			Timers:   gostatsd.Timers{},
			Gauges:   gostatsd.Gauges{},
			Sets:     gostatsd.Sets{},
		},
		disabledSubtypes: disabled,
	}
	for _, pct := range percentThresholds {
		sPct := strconv.Itoa(int(pct))
		a.percentThresholds[pct] = percentStruct{
			count:      "count_" + sPct,
			mean:       "mean_" + sPct,
			sum:        "sum_" + sPct,
			sumSquares: "sum_squares_" + sPct,
			upper:      "upper_" + sPct,
			lower:      "lower_" + sPct,
		}
	}
	return &a
}

// round rounds a number to its nearest integer value.
// poor man's math.Round(x) = math.Floor(x + 0.5).
func round(v float64) float64 {
	return math.Floor(v + 0.5)
}

// Flush prepares the contents of a MetricAggregator for sending via the Sender.
func (a *MetricAggregator) Flush(flushInterval time.Duration) {
	a.statser.Gauge("aggregator.metrics_received", float64(a.metricsReceived), nil)

	flushInSeconds := float64(flushInterval) / float64(time.Second)

	a.Counters.Each(func(key, tagsKey string, counter gostatsd.Counter) {
		counter.PerSecond = float64(counter.Value) / flushInSeconds
		a.Counters[key][tagsKey] = counter
	})

	a.Timers.Each(func(key, tagsKey string, timer gostatsd.Timer) {
		if count := len(timer.Values); count > 0 {
			sort.Float64s(timer.Values)
			timer.Min = timer.Values[0]
			timer.Max = timer.Values[count-1]
			timer.Count = len(timer.Values)
			count := float64(timer.Count)

			cumulativeValues := make([]float64, timer.Count)
			cumulSumSquaresValues := make([]float64, timer.Count)
			cumulativeValues[0] = timer.Min
			cumulSumSquaresValues[0] = timer.Min * timer.Min
			for i := 1; i < timer.Count; i++ {
				cumulativeValues[i] = timer.Values[i] + cumulativeValues[i-1]
				cumulSumSquaresValues[i] = timer.Values[i]*timer.Values[i] + cumulSumSquaresValues[i-1]
			}

			var sumSquares = timer.Min * timer.Min
			var mean = timer.Min
			var sum = timer.Min
			var thresholdBoundary = timer.Max

			for pct, pctStruct := range a.percentThresholds {
				numInThreshold := timer.Count
				if timer.Count > 1 {
					numInThreshold = int(round(math.Abs(pct) / 100 * count))
					if numInThreshold == 0 {
						continue
					}
					if pct > 0 {
						thresholdBoundary = timer.Values[numInThreshold-1]
						sum = cumulativeValues[numInThreshold-1]
						sumSquares = cumulSumSquaresValues[numInThreshold-1]
					} else {
						thresholdBoundary = timer.Values[timer.Count-numInThreshold]
						sum = cumulativeValues[timer.Count-1] - cumulativeValues[timer.Count-numInThreshold-1]
						sumSquares = cumulSumSquaresValues[timer.Count-1] - cumulSumSquaresValues[timer.Count-numInThreshold-1]
					}
					mean = sum / float64(numInThreshold)
				}

				if !a.disabledSubtypes.CountPct {
					timer.Percentiles.Set(pctStruct.count, float64(numInThreshold))
				}
				if !a.disabledSubtypes.MeanPct {
					timer.Percentiles.Set(pctStruct.mean, mean)
				}
				if !a.disabledSubtypes.SumPct {
					timer.Percentiles.Set(pctStruct.sum, sum)
				}
				if !a.disabledSubtypes.SumSquaresPct {
					timer.Percentiles.Set(pctStruct.sumSquares, sumSquares)
				}
				if pct > 0 {
					if !a.disabledSubtypes.UpperPct {
						timer.Percentiles.Set(pctStruct.upper, thresholdBoundary)
					}
				} else {
					if !a.disabledSubtypes.LowerPct {
						timer.Percentiles.Set(pctStruct.lower, thresholdBoundary)
					}
				}
			}

			sum = cumulativeValues[timer.Count-1]
			sumSquares = cumulSumSquaresValues[timer.Count-1]
			mean = sum / count

			var sumOfDiffs float64
			for i := 0; i < timer.Count; i++ {
				sumOfDiffs += (timer.Values[i] - mean) * (timer.Values[i] - mean)
			}

			mid := int(math.Floor(count / 2))
			if math.Mod(count, 2) == 0 {
				timer.Median = (timer.Values[mid-1] + timer.Values[mid]) / 2
			} else {
				timer.Median = timer.Values[mid]
			}

			timer.Mean = mean
			timer.StdDev = math.Sqrt(sumOfDiffs / count)
			timer.Sum = sum
			timer.SumSquares = sumSquares
			timer.PerSecond = count / flushInSeconds

			a.Timers[key][tagsKey] = timer
		} else {
			timer.Count = 0
			timer.PerSecond = 0
		}
	})
}

func (a *MetricAggregator) RunMetrics(ctx context.Context, statser statser.Statser) {
	a.statser = statser
}

func (a *MetricAggregator) Process(f ProcessFunc) {
	f(&a.MetricMap)
}

func (a *MetricAggregator) isExpired(now, ts gostatsd.Nanotime) bool {
	return a.expiryInterval != 0 && time.Duration(now-ts) > a.expiryInterval
}

func deleteMetric(key, tagsKey string, metrics gostatsd.AggregatedMetrics) {
	metrics.DeleteChild(key, tagsKey)
	if !metrics.HasChildren(key) {
		metrics.Delete(key)
	}
}

// Reset clears the contents of a MetricAggregator.
func (a *MetricAggregator) Reset() {
	a.metricsReceived = 0
	nowNano := gostatsd.Nanotime(a.now().UnixNano())

	a.Counters.Each(func(key, tagsKey string, counter gostatsd.Counter) {
		if a.isExpired(nowNano, counter.Timestamp) {
			deleteMetric(key, tagsKey, a.Counters)
		} else {
			a.Counters[key][tagsKey] = gostatsd.Counter{
				Timestamp: counter.Timestamp,
				Hostname:  counter.Hostname,
				Tags:      counter.Tags,
			}
		}
	})

	a.Timers.Each(func(key, tagsKey string, timer gostatsd.Timer) {
		if a.isExpired(nowNano, timer.Timestamp) {
			deleteMetric(key, tagsKey, a.Timers)
		} else {
			a.Timers[key][tagsKey] = gostatsd.Timer{
				Timestamp: timer.Timestamp,
				Hostname:  timer.Hostname,
				Tags:      timer.Tags,
				Values:    timer.Values[:0],
			}
		}
	})

	a.Gauges.Each(func(key, tagsKey string, gauge gostatsd.Gauge) {
		if a.isExpired(nowNano, gauge.Timestamp) {
			deleteMetric(key, tagsKey, a.Gauges)
		}
		// No reset for gauges, they keep the last value until expiration
	})

	a.Sets.Each(func(key, tagsKey string, set gostatsd.Set) {
		if a.isExpired(nowNano, set.Timestamp) {
			deleteMetric(key, tagsKey, a.Sets)
		} else {
			a.Sets[key][tagsKey] = gostatsd.Set{
				Values:    make(map[string]struct{}),
				Timestamp: set.Timestamp,
				Hostname:  set.Hostname,
				Tags:      set.Tags,
			}
		}
	})
}

func (a *MetricAggregator) receiveCounter(m *gostatsd.Metric, tagsKey string, now gostatsd.Nanotime) {
	value := int64(m.Value)
	v, ok := a.Counters[m.Name]
	if ok {
		c, ok := v[tagsKey]
		if ok {
			c.Value += value
			c.Timestamp = now
		} else {
			c = gostatsd.NewCounter(now, value, m.Hostname, m.Tags)
		}
		v[tagsKey] = c
	} else {
		a.Counters[m.Name] = map[string]gostatsd.Counter{
			tagsKey: gostatsd.NewCounter(now, value, m.Hostname, m.Tags),
		}
	}
}

func (a *MetricAggregator) receiveGauge(m *gostatsd.Metric, tagsKey string, now gostatsd.Nanotime) {
	// TODO: handle +/-
	v, ok := a.Gauges[m.Name]
	if ok {
		g, ok := v[tagsKey]
		if ok {
			g.Value = m.Value
			g.Timestamp = now
		} else {
			g = gostatsd.NewGauge(now, m.Value, m.Hostname, m.Tags)
		}
		v[tagsKey] = g
	} else {
		a.Gauges[m.Name] = map[string]gostatsd.Gauge{
			tagsKey: gostatsd.NewGauge(now, m.Value, m.Hostname, m.Tags),
		}
	}
}

func (a *MetricAggregator) receiveTimer(m *gostatsd.Metric, tagsKey string, now gostatsd.Nanotime) {
	v, ok := a.Timers[m.Name]
	if ok {
		t, ok := v[tagsKey]
		if ok {
			t.Values = append(t.Values, m.Value)
			t.Timestamp = now
		} else {
			t = gostatsd.NewTimer(now, []float64{m.Value}, m.Hostname, m.Tags)
		}
		v[tagsKey] = t
	} else {
		a.Timers[m.Name] = map[string]gostatsd.Timer{
			tagsKey: gostatsd.NewTimer(now, []float64{m.Value}, m.Hostname, m.Tags),
		}
	}
}

func (a *MetricAggregator) receiveSet(m *gostatsd.Metric, tagsKey string, now gostatsd.Nanotime) {
	v, ok := a.Sets[m.Name]
	if ok {
		s, ok := v[tagsKey]
		if ok {
			s.Values[m.StringValue] = struct{}{}
			s.Timestamp = now
		} else {
			s = gostatsd.NewSet(now, map[string]struct{}{m.StringValue: {}}, m.Hostname, m.Tags)
		}
		v[tagsKey] = s
	} else {
		a.Sets[m.Name] = map[string]gostatsd.Set{
			tagsKey: gostatsd.NewSet(now, map[string]struct{}{m.StringValue: {}}, m.Hostname, m.Tags),
		}
	}
}

// Receive aggregates an incoming metric.
func (a *MetricAggregator) Receive(m *gostatsd.Metric, now time.Time) {
	a.metricsReceived++
	tagsKey := m.TagsKey
	nowNano := gostatsd.Nanotime(now.UnixNano())

	switch m.Type {
	case gostatsd.COUNTER:
		a.receiveCounter(m, tagsKey, nowNano)
	case gostatsd.GAUGE:
		a.receiveGauge(m, tagsKey, nowNano)
	case gostatsd.TIMER:
		a.receiveTimer(m, tagsKey, nowNano)
	case gostatsd.SET:
		a.receiveSet(m, tagsKey, nowNano)
	default:
		log.Errorf("Unknow metric type %s for %s", m.Type, m.Name)
	}
	m.Done()
}
