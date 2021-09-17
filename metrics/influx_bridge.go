package metrics

import (
	"context"
	"errors"
	"io"
	"time"

	influx "github.com/influxdata/line-protocol"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

// InfluxBridge is a bridge between prometheus metrics and influxdb line metrics,
// similar to the graphite bridge published by prometheus.
// https://github.com/prometheus/client_golang/blob/v1.11.0/prometheus/graphite
//
// See here for details about how to transform prometheus metrics into influx metrics:
// https://www.influxdata.com/blog/prometheus-remote-write-support-with-influxdb-2-0/
type InfluxBridge struct {
	interval  time.Duration
	gatherer  prometheus.Gatherer
	writer    io.Writer
	errLogger *logrus.Entry
	filter    map[string]struct{}
}

// InfluxBridgeConfig configures an InfluxBridge.
type InfluxBridgeConfig struct {
	// Writer specifies the writer to use. Required.
	Writer io.Writer

	// Interval specifies the logging interval to use. Required.
	Interval time.Duration

	// Gatherer specifies the prometheus gatherer to get metrics from. Required.
	Gatherer prometheus.Gatherer

	// ErrLogger specifies the logrus logger to use. Set to a default if not
	// supplied.
	ErrLogger *logrus.Entry

	// Select, if non-nil, will limit the exported metrics to the names present
	// in the list.
	Select []string
}

// NewInfluxBridge creates a new InfluxBridge. If the supplied InfluxBridgeConfig
// is not correctly formed, an error will be returned.
func NewInfluxBridge(cfg *InfluxBridgeConfig) (*InfluxBridge, error) {
	bridge := &InfluxBridge{
		filter: make(map[string]struct{}),
	}
	for _, selectedMetric := range cfg.Select {
		bridge.filter[selectedMetric] = struct{}{}
	}
	if cfg.Interval == 0 {
		return nil, errors.New("influx bridge interval not set")
	}
	bridge.interval = cfg.Interval
	if bridge.interval < time.Second {
		bridge.interval = time.Second * bridge.interval
	}
	if cfg.Gatherer == nil {
		return nil, errors.New("nil gatherer")
	}
	bridge.gatherer = cfg.Gatherer
	bridge.writer = cfg.Writer
	bridge.errLogger = cfg.ErrLogger
	if bridge.errLogger == nil {
		bridge.errLogger = logrus.NewEntry(logrus.StandardLogger())
	}
	return bridge, nil
}

// Run starts the bridge. It operates in a blocking fashion, and runs until the
// supplied context is cancelled.
func (b *InfluxBridge) Run(ctx context.Context) {
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := b.Push(); err != nil {
				b.errLogger.WithError(err).Error("error logging platform metrics")
			}
		}
	}
}

// Push pushes the current set of gathered metrics to b's writer.
func (b *InfluxBridge) Push() error {
	metrics, err := b.gatherer.Gather()
	if err != nil {
		return err
	}
	return b.logMetrics(metrics)
}

type promSampleInfluxMetric model.Sample

func (p *promSampleInfluxMetric) Time() time.Time {
	nanos := int64(p.Timestamp) * int64(time.Millisecond)
	return time.Unix(0, nanos)
}

func (p *promSampleInfluxMetric) Name() string {
	name := string(p.Metric[model.MetricNameLabel])
	return name
}

func (p *promSampleInfluxMetric) TagList() []*influx.Tag {
	tags := []*influx.Tag{}
	for k, v := range p.Metric {
		if k == model.MetricNameLabel {
			continue
		}
		tags = append(tags, &influx.Tag{Key: string(k), Value: string(v)})
	}
	return tags
}

func (p *promSampleInfluxMetric) FieldList() []*influx.Field {
	fields := []*influx.Field{
		&influx.Field{
			Key:   p.Name(),
			Value: float64(p.Value),
		},
	}
	return fields
}

func (b *InfluxBridge) logMetrics(families []*dto.MetricFamily) error {
	now := model.Now() // milliseconds since the epoch, excluding leap seconds
	samples, err := expfmt.ExtractSamples(&expfmt.DecodeOptions{Timestamp: now}, families...)
	if err != nil {
		// some metrics might have been successfully extracted, soldier on
		b.errLogger.WithError(err).Error("error extracting prometheus metric samples")
	}

	encoder := influx.NewEncoder(b.writer)
	encoder.FailOnFieldErr(true)

	for _, sample := range samples {
		metric := (*promSampleInfluxMetric)(sample)
		if len(b.filter) > 0 {
			if _, ok := b.filter[metric.Name()]; !ok {
				continue
			}
		}
		if _, err := encoder.Encode(metric); err != nil {
			b.errLogger.WithError(err).Error("error encoding metric")
		}
	}

	return err
}
