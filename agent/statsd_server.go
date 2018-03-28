// Package agent is the running Sensu agent. Agents connect to a Sensu backend,
// register their presence, subscribe to check channels, download relevant
// check packages, execute checks, and send results to the Sensu backend via
// the Event channel.
package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/atlassian/gostatsd"
	"github.com/atlassian/gostatsd/pkg/statsd"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/viper"
)

// NewStatsdServer provides a new statsd server for the sensu-agent.
func NewStatsdServer(c *StatsdServerConfig) *statsd.Server {
	s := statsd.NewServer()
	backend, err := NewClientFromViper(s.Viper)
	if err != nil {
		logger.WithError(err).Error("failed to create sensu statsd backend")
	}
	s.Backends = []gostatsd.Backend{backend}
	s.FlushInterval = time.Duration(c.FlushInterval) * time.Second
	s.MetricsAddr = fmt.Sprintf("%s:%d", c.Host, c.Port)
	return s
}

// BackendName is the name of this statsd backend.
const BackendName = "sensu-statsd"

// Client is an object that is used to send messages to sensu-statsd.
type Client struct{}

// NewClientFromViper constructs a sensu statsd backend.
func NewClientFromViper(v *viper.Viper) (gostatsd.Backend, error) {
	return NewClient()
}

// NewClient constructs a sensu-statsd backend.
func NewClient() (*Client, error) {
	return &Client{}, nil
}

// SendMetricsAsync flushes the metrics to the statsd backend which resides on
// the sensu-agent, preparing payload synchronously but doing the send asynchronously.
// Must not read/write MetricMap asynchronously.
func (client Client) SendMetricsAsync(ctx context.Context, metrics *gostatsd.MetricMap, cb gostatsd.SendCallback) {
	metricsPoints := prepareMetrics(metrics)
	go func() {
		cb([]error{sendMetrics(metricsPoints)})
	}()
}

// SendEvent sends event to the statsd backend which resides on the sensu-agent,
// not to be confused with the sensu-backend.
func (Client) SendEvent(ctx context.Context, e *gostatsd.Event) (retErr error) {
	logger.WithField("event", e).Info("statsd received an event")
	return nil
}

// Name returns the name of the backend.
func (Client) Name() string {
	return BackendName
}

func prepareMetrics(metrics *gostatsd.MetricMap) []*types.MetricPoint {
	now := time.Now().Unix()
	var metricsPoints []*types.MetricPoint
	metrics.Counters.Each(func(key, tagsKey string, counter gostatsd.Counter) {
		tags := composeMetricTags(tagsKey)
		counters := composeCounterPoints(counter, key, tags, now)
		metricsPoints = append(metricsPoints, counters...)
	})
	metrics.Timers.Each(func(key, tagsKey string, timer gostatsd.Timer) {
		tags := composeMetricTags(tagsKey)
		timers := composeTimerPoints(timer, key, tags, now)
		metricsPoints = append(metricsPoints, timers...)
	})
	metrics.Gauges.Each(func(key, tagsKey string, gauge gostatsd.Gauge) {
		tags := composeMetricTags(tagsKey)
		gauges := composeGaugePoints(gauge, key, tags, now)
		metricsPoints = append(metricsPoints, gauges...)
	})
	metrics.Sets.Each(func(key, tagsKey string, set gostatsd.Set) {
		tags := composeMetricTags(tagsKey)
		sets := composeSetPoints(set, key, tags, now)
		metricsPoints = append(metricsPoints, sets...)
	})
	return metricsPoints
}

func sendMetrics(metrics []*types.MetricPoint) (retErr error) {
	for _, metric := range metrics {
		logger.WithField("metric", metric).Info("metric received from statsd")
	}
	return nil
}

func composeMetricTags(tagsKey string) []*types.MetricTag {
	tagsKeys := strings.Split(tagsKey, ",")
	var tags []*types.MetricTag
	var name, value string
	for _, tag := range tagsKeys {
		tagsValues := strings.Split(tag, ":")
		if len(tagsValues) > 1 {
			name = tagsValues[0]
			value = tagsValues[1]
		}
		if tag != "" {
			t := &types.MetricTag{
				Name:  name,
				Value: value,
			}
			tags = append(tags, t)
		}
	}
	return tags
}

func composeCounterPoints(counter gostatsd.Counter, key string, tags []*types.MetricTag, now int64) []*types.MetricPoint {
	m0 := &types.MetricPoint{
		Name:      key + ".value",
		Value:     float64(counter.Value),
		Timestamp: now,
		Tags:      tags,
	}
	m1 := &types.MetricPoint{
		Name:      key + ".per_second",
		Value:     float64(counter.PerSecond),
		Timestamp: now,
		Tags:      tags,
	}
	points := []*types.MetricPoint{m0, m1}
	return points
}

func composeTimerPoints(timer gostatsd.Timer, key string, tags []*types.MetricTag, now int64) []*types.MetricPoint {
	m0 := &types.MetricPoint{
		Name:      key + ".min",
		Value:     timer.Min,
		Timestamp: now,
		Tags:      tags,
	}
	m1 := &types.MetricPoint{
		Name:      key + ".max",
		Value:     timer.Max,
		Timestamp: now,
		Tags:      tags,
	}
	m2 := &types.MetricPoint{
		Name:      key + ".count",
		Value:     float64(timer.Count),
		Timestamp: now,
		Tags:      tags,
	}
	m3 := &types.MetricPoint{
		Name:      key + ".per_second",
		Value:     timer.PerSecond,
		Timestamp: now,
		Tags:      tags,
	}
	m4 := &types.MetricPoint{
		Name:      key + ".mean",
		Value:     timer.Mean,
		Timestamp: now,
		Tags:      tags,
	}
	m5 := &types.MetricPoint{
		Name:      key + ".median",
		Value:     timer.Median,
		Timestamp: now,
		Tags:      tags,
	}
	m6 := &types.MetricPoint{
		Name:      key + ".stddev",
		Value:     timer.StdDev,
		Timestamp: now,
		Tags:      tags,
	}
	m7 := &types.MetricPoint{
		Name:      key + ".sum",
		Value:     timer.Sum,
		Timestamp: now,
		Tags:      tags,
	}
	m8 := &types.MetricPoint{
		Name:      key + ".sum_squares",
		Value:     timer.SumSquares,
		Timestamp: now,
		Tags:      tags,
	}
	points := []*types.MetricPoint{m0, m1, m2, m3, m4, m5, m6, m7, m8}
	for _, pct := range timer.Percentiles {
		m := &types.MetricPoint{
			Name:      key + ".percentile_" + pct.Str,
			Value:     pct.Float,
			Timestamp: now,
			Tags:      tags,
		}
		points = append(points, m)
	}
	return points
}

func composeGaugePoints(gauge gostatsd.Gauge, key string, tags []*types.MetricTag, now int64) []*types.MetricPoint {
	m0 := &types.MetricPoint{
		Name:      key + ".value",
		Value:     float64(gauge.Value),
		Timestamp: now,
		Tags:      tags,
	}
	points := []*types.MetricPoint{m0}
	return points
}

func composeSetPoints(set gostatsd.Set, key string, tags []*types.MetricTag, now int64) []*types.MetricPoint {
	m0 := &types.MetricPoint{
		Name:      key + ".value",
		Value:     float64(len(set.Values)),
		Timestamp: now,
		Tags:      tags,
	}
	points := []*types.MetricPoint{m0}
	return points
}
