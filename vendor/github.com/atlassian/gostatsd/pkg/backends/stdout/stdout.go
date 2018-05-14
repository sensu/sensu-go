package stdout

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/atlassian/gostatsd"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// BackendName is the name of this backend.
const BackendName = "stdout"

// Client is an object that is used to send messages to stdout.
type Client struct {
	disabledSubtypes gostatsd.TimerSubtypes
}

// NewClientFromViper constructs a stdout backend.
func NewClientFromViper(v *viper.Viper) (gostatsd.Backend, error) {
	return NewClient(
		gostatsd.DisabledSubMetrics(v),
	)
}

// NewClient constructs a stdout backend.
func NewClient(disabled gostatsd.TimerSubtypes) (*Client, error) {
	return &Client{
		disabledSubtypes: disabled,
	}, nil
}

// composeMetricName adds the key and the tags to compose the metric name.
func composeMetricName(key string, tagsKey string) string {
	tags := strings.Split(tagsKey, ",")
	for _, tag := range tags {
		if tag != "" {
			key += "." + tagToMetricName(tag)
		}
	}
	return key
}

// tagToMetricName transforms tags into metric names.
func tagToMetricName(tag string) string {
	return strings.Replace(tag, ":", ".", -1)
}

// SendMetricsAsync prints the metrics in a MetricsMap to the stdout, preparing payload synchronously but doing the send asynchronously.
func (client Client) SendMetricsAsync(ctx context.Context, metrics *gostatsd.MetricMap, cb gostatsd.SendCallback) {
	buf := preparePayload(metrics, &client.disabledSubtypes)
	go func() {
		cb([]error{writePayload(buf)})
	}()
}

func writePayload(buf *bytes.Buffer) (retErr error) {
	writer := log.StandardLogger().Writer()
	defer func() {
		if err := writer.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	_, err := writer.Write(buf.Bytes())
	return err
}

func preparePayload(metrics *gostatsd.MetricMap, disabled *gostatsd.TimerSubtypes) *bytes.Buffer {
	buf := new(bytes.Buffer)
	now := time.Now().Unix()
	metrics.Counters.Each(func(key, tagsKey string, counter gostatsd.Counter) {
		nk := composeMetricName(key, tagsKey)
		fmt.Fprintf(buf, "stats.counter.%s.count %d %d\n", nk, counter.Value, now)          // #nosec
		fmt.Fprintf(buf, "stats.counter.%s.per_second %f %d\n", nk, counter.PerSecond, now) // #nosec
	})
	metrics.Timers.Each(func(key, tagsKey string, timer gostatsd.Timer) {
		nk := composeMetricName(key, tagsKey)
		if !disabled.Lower {
			fmt.Fprintf(buf, "stats.timers.%s.lower %f %d\n", nk, timer.Min, now) // #nosec
		}
		if !disabled.Upper {
			fmt.Fprintf(buf, "stats.timers.%s.upper %f %d\n", nk, timer.Max, now) // #nosec
		}
		if !disabled.Count {
			fmt.Fprintf(buf, "stats.timers.%s.count %d %d\n", nk, timer.Count, now) // #nosec
		}
		if !disabled.CountPerSecond {
			fmt.Fprintf(buf, "stats.timers.%s.count_ps %f %d\n", nk, timer.PerSecond, now) // #nosec
		}
		if !disabled.Mean {
			fmt.Fprintf(buf, "stats.timers.%s.mean %f %d\n", nk, timer.Mean, now) // #nosec
		}
		if !disabled.Median {
			fmt.Fprintf(buf, "stats.timers.%s.median %f %d\n", nk, timer.Median, now) // #nosec
		}
		if !disabled.StdDev {
			fmt.Fprintf(buf, "stats.timers.%s.std %f %d\n", nk, timer.StdDev, now) // #nosec
		}
		if !disabled.Sum {
			fmt.Fprintf(buf, "stats.timers.%s.sum %f %d\n", nk, timer.Sum, now) // #nosec
		}
		if !disabled.SumSquares {
			fmt.Fprintf(buf, "stats.timers.%s.sum_squares %f %d\n", nk, timer.SumSquares, now) // #nosec
		}
		for _, pct := range timer.Percentiles {
			fmt.Fprintf(buf, "stats.timers.%s.%s %f %d\n", nk, pct.Str, pct.Float, now) // #nosec
		}
	})
	metrics.Gauges.Each(func(key, tagsKey string, gauge gostatsd.Gauge) {
		nk := composeMetricName(key, tagsKey)
		fmt.Fprintf(buf, "stats.gauge.%s %f %d\n", nk, gauge.Value, now) // #nosec
	})
	metrics.Sets.Each(func(key, tagsKey string, set gostatsd.Set) {
		nk := composeMetricName(key, tagsKey)
		fmt.Fprintf(buf, "stats.set.%s %d %d\n", nk, len(set.Values), now) // #nosec
	})
	return buf
}

// SendEvent prints events to the stdout.
func (client Client) SendEvent(ctx context.Context, e *gostatsd.Event) (retErr error) {
	writer := log.StandardLogger().Writer()
	defer func() {
		if err := writer.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	_, err := fmt.Fprintf(writer, "event: %+v\n", e)
	return err
}

// Name returns the name of the backend.
func (Client) Name() string {
	return BackendName
}
