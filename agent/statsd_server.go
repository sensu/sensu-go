//go:build !solaris
// +build !solaris

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
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/transport"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

// GetMetricsAddr gets the metrics address of the statsd server.
func GetMetricsAddr(s StatsdServer) string {
	server, ok := s.(*statsd.Server)
	if !ok {
		return ""
	}
	return server.MetricsAddr
}

// NewStatsdServer provides a new statsd server for the sensu-agent.
func NewStatsdServer(a *Agent) *statsd.Server {
	c := a.config.StatsdServer
	s := NewServer()
	backend, err := NewClientFromViper(s.Viper, a)
	if err != nil {
		logger.WithError(err).Error("failed to create sensu-statsd backend")
	}
	s.Backends = []gostatsd.Backend{backend}
	if c.FlushInterval == 0 {
		logger.Error("invalid statsd flush interval of 0, using the default 10s")
		c.FlushInterval = DefaultStatsdFlushInterval
	}
	s.FlushInterval = time.Duration(c.FlushInterval) * time.Second
	s.MetricsAddr = fmt.Sprintf("%s:%d", c.Host, c.Port)
	s.StatserType = statsd.StatserNull
	return s
}

// NewServer will create a new statsd Server with the default configuration.
func NewServer() *statsd.Server {
	return &statsd.Server{
		Backends:            []gostatsd.Backend{},
		Limiter:             rate.NewLimiter(statsd.DefaultMaxCloudRequests, statsd.DefaultBurstCloudRequests),
		InternalTags:        statsd.DefaultInternalTags,
		InternalNamespace:   statsd.DefaultInternalNamespace,
		DefaultTags:         statsd.DefaultTags,
		ExpiryInterval:      statsd.DefaultExpiryInterval,
		FlushInterval:       statsd.DefaultFlushInterval,
		MaxReaders:          statsd.DefaultMaxReaders,
		MaxParsers:          statsd.DefaultMaxParsers,
		MaxWorkers:          statsd.DefaultMaxWorkers,
		MaxQueueSize:        statsd.DefaultMaxQueueSize,
		MaxConcurrentEvents: statsd.DefaultMaxConcurrentEvents,
		EstimatedTags:       statsd.DefaultEstimatedTags,
		MetricsAddr:         statsd.DefaultMetricsAddr,
		PercentThreshold:    statsd.DefaultPercentThreshold,
		IgnoreHost:          statsd.DefaultIgnoreHost,
		ConnPerReader:       statsd.DefaultConnPerReader,
		HeartbeatEnabled:    statsd.DefaultHeartbeatEnabled,
		ReceiveBatchSize:    statsd.DefaultReceiveBatchSize,
		CacheOptions: statsd.CacheOptions{
			CacheRefreshPeriod:        statsd.DefaultCacheRefreshPeriod,
			CacheEvictAfterIdlePeriod: statsd.DefaultCacheEvictAfterIdlePeriod,
			CacheTTL:                  statsd.DefaultCacheTTL,
			CacheNegativeTTL:          statsd.DefaultCacheNegativeTTL,
		},
		Viper: viper.New(),
	}
}

// BackendName is the name of this statsd backend.
const BackendName = "sensu-statsd"

// Client is an object that is used to send messages to sensu-statsd.
type Client struct {
	agent *Agent
}

// NewClientFromViper constructs a sensu-statsd backend.
func NewClientFromViper(v *viper.Viper, a *Agent) (gostatsd.Backend, error) {
	return NewClient(a)
}

// NewClient constructs a sensu-statsd backend.
func NewClient(a *Agent) (*Client, error) {
	return &Client{agent: a}, nil
}

// SendMetricsAsync flushes the metrics to the statsd backend which resides on
// the sensu-agent, preparing payload synchronously but doing the send asynchronously.
// Must not read/write MetricMap asynchronously.
func (c Client) SendMetricsAsync(ctx context.Context, metrics *gostatsd.MetricMap, cb gostatsd.SendCallback) {
	now := time.Now().UnixNano()
	metricsPoints := prepareMetrics(now, metrics)
	go func() {
		cb([]error{c.sendMetrics(metricsPoints)})
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

func prepareMetrics(now int64, metrics *gostatsd.MetricMap) []*corev2.MetricPoint {
	var metricsPoints []*corev2.MetricPoint
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

func (c Client) sendMetrics(points []*corev2.MetricPoint) (retErr error) {
	if points == nil {
		return nil
	}

	metrics := &corev2.Metrics{
		Points:   points,
		Handlers: c.agent.config.StatsdServer.Handlers,
	}
	event := &corev2.Event{
		Entity:    c.agent.getAgentEntity(),
		Timestamp: time.Now().Unix(),
		Metrics:   metrics,
	}

	msg, err := c.agent.marshal(event)
	if err != nil {
		logger.WithError(err).Error("error marshaling metric event")
		return err
	}

	logger.WithFields(logrus.Fields{
		"metrics": event.Metrics,
		"entity":  event.Entity.Name,
	}).Debug("sending statsd metrics")
	tm := &transport.Message{
		Type:    transport.MessageTypeEvent,
		Payload: msg,
	}
	c.agent.sendMessage(tm)
	return nil
}

func composeMetricTags(tagsKey string) []*corev2.MetricTag {
	tagsKeys := strings.Split(tagsKey, ",")
	var tags []*corev2.MetricTag
	var name, value string
	for _, tag := range tagsKeys {
		tagsValues := strings.Split(tag, ":")
		if len(tagsValues) > 1 {
			name = tagsValues[0]
			value = tagsValues[1]
		}
		if tag != "" {
			t := &corev2.MetricTag{
				Name:  name,
				Value: value,
			}
			tags = append(tags, t)
		}
	}
	return tags
}

func composeCounterPoints(counter gostatsd.Counter, key string, tags []*corev2.MetricTag, now int64) []*corev2.MetricPoint {
	m0 := &corev2.MetricPoint{
		Name:      key + ".value",
		Value:     float64(counter.Value),
		Timestamp: now,
		Tags:      tags,
	}
	m1 := &corev2.MetricPoint{
		Name:      key + ".per_second",
		Value:     float64(counter.PerSecond),
		Timestamp: now,
		Tags:      tags,
	}
	points := []*corev2.MetricPoint{m0, m1}
	return points
}

func composeTimerPoints(timer gostatsd.Timer, key string, tags []*corev2.MetricTag, now int64) []*corev2.MetricPoint {
	m0 := &corev2.MetricPoint{
		Name:      key + ".min",
		Value:     timer.Min,
		Timestamp: now,
		Tags:      tags,
	}
	m1 := &corev2.MetricPoint{
		Name:      key + ".max",
		Value:     timer.Max,
		Timestamp: now,
		Tags:      tags,
	}
	m2 := &corev2.MetricPoint{
		Name:      key + ".count",
		Value:     float64(timer.Count),
		Timestamp: now,
		Tags:      tags,
	}
	m3 := &corev2.MetricPoint{
		Name:      key + ".per_second",
		Value:     timer.PerSecond,
		Timestamp: now,
		Tags:      tags,
	}
	m4 := &corev2.MetricPoint{
		Name:      key + ".mean",
		Value:     timer.Mean,
		Timestamp: now,
		Tags:      tags,
	}
	m5 := &corev2.MetricPoint{
		Name:      key + ".median",
		Value:     timer.Median,
		Timestamp: now,
		Tags:      tags,
	}
	m6 := &corev2.MetricPoint{
		Name:      key + ".stddev",
		Value:     timer.StdDev,
		Timestamp: now,
		Tags:      tags,
	}
	m7 := &corev2.MetricPoint{
		Name:      key + ".sum",
		Value:     timer.Sum,
		Timestamp: now,
		Tags:      tags,
	}
	m8 := &corev2.MetricPoint{
		Name:      key + ".sum_squares",
		Value:     timer.SumSquares,
		Timestamp: now,
		Tags:      tags,
	}
	points := []*corev2.MetricPoint{m0, m1, m2, m3, m4, m5, m6, m7, m8}
	for _, pct := range timer.Percentiles {
		m := &corev2.MetricPoint{
			Name:      key + ".percentile_" + pct.Str,
			Value:     pct.Float,
			Timestamp: now,
			Tags:      tags,
		}
		points = append(points, m)
	}
	return points
}

func composeGaugePoints(gauge gostatsd.Gauge, key string, tags []*corev2.MetricTag, now int64) []*corev2.MetricPoint {
	m0 := &corev2.MetricPoint{
		Name:      key + ".value",
		Value:     float64(gauge.Value),
		Timestamp: now,
		Tags:      tags,
	}
	points := []*corev2.MetricPoint{m0}
	return points
}

func composeSetPoints(set gostatsd.Set, key string, tags []*corev2.MetricTag, now int64) []*corev2.MetricPoint {
	m0 := &corev2.MetricPoint{
		Name:      key + ".value",
		Value:     float64(len(set.Values)),
		Timestamp: now,
		Tags:      tags,
	}
	points := []*corev2.MetricPoint{m0}
	return points
}
