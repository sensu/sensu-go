package transformers

import (
	"math"
	"testing"

	time "github.com/echlebek/timeproxy"

	"github.com/prometheus/common/model"
	v2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
)

func TestParseProm(t *testing.T) {
	assert := assert.New(t)
	ts := time.Now().Unix()

	testCases := []struct {
		metric           string
		expectedFormat   PromList
		timeInconclusive bool
	}{
		{
			metric: "go_gc_duration_seconds{quantile=\"0\"} 3.3722e-05\n",
			expectedFormat: PromList{
				&model.Sample{
					Metric: model.Metric{
						model.MetricNameLabel: "go_gc_duration_seconds",
						"quantile":            "0",
						"prom_type":           "untyped",
					},
					Value:     3.3722e-05,
					Timestamp: model.TimeFromUnix(ts),
				},
			},
		},
		{
			metric: "# TYPE go_gc_duration_seconds summary\ngo_gc_duration_seconds{quantile=\"0\"} 3.3722e-05\ngo_gc_duration_seconds{quantile=\"0.25\"} 5.0129e-05\n",
			expectedFormat: PromList{
				&model.Sample{
					Metric: model.Metric{
						model.MetricNameLabel: "go_gc_duration_seconds",
						"quantile":            "0",
						"prom_type":           "summary",
					},
					Value:     3.3722e-05,
					Timestamp: model.TimeFromUnix(ts),
				},
				&model.Sample{
					Metric: model.Metric{
						model.MetricNameLabel: "go_gc_duration_seconds",
						"quantile":            "0.25",
						"prom_type":           "summary",
					},
					Value:     5.0129e-05,
					Timestamp: model.TimeFromUnix(ts),
				},
				&model.Sample{
					Metric: model.Metric{
						model.MetricNameLabel: "go_gc_duration_seconds_sum",
						"prom_type":           "summary",
					},
					Value:     0,
					Timestamp: model.TimeFromUnix(ts),
				},
				&model.Sample{
					Metric: model.Metric{
						model.MetricNameLabel: "go_gc_duration_seconds_count",
						"prom_type":           "summary",
					},
					Value:     0,
					Timestamp: model.TimeFromUnix(ts),
				},
			},
		},
		{
			metric: "# HELP go_memstats_alloc_bytes_total Total number of bytes allocated, even if freed.\n# TYPE go_memstats_alloc_bytes_total counter\ngo_memstats_alloc_bytes_total 4.095146016e+09\n",
			expectedFormat: PromList{
				&model.Sample{
					Metric: model.Metric{
						model.MetricNameLabel: "go_memstats_alloc_bytes_total",
						"prom_type":           "counter",
						"prom_help":           "Total number of bytes allocated, even if freed.",
					},
					Value:     4.095146016e+09,
					Timestamp: model.TimeFromUnix(ts),
				},
			},
		},
		{
			metric:         "foo 1",
			expectedFormat: PromList{},
		},
		{
			metric:         "foo{bar=\"2\"}\n",
			expectedFormat: PromList{},
		},
		{
			metric:         "",
			expectedFormat: PromList{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			event := v2.FixtureEvent("test", "test")
			event.Check.Output = tc.metric
			prom := ParseProm(event)
			if !tc.timeInconclusive {
				assert.Equal(tc.expectedFormat, prom)
			}
		})
	}
}

func TestParsePromTags(t *testing.T) {
	assert := assert.New(t)
	ts := time.Now().Unix()

	testCases := []struct {
		metric           string
		expectedFormat   PromList
		timeInconclusive bool
		outputMetricTags []*v2.MetricTag
	}{
		{
			metric: "go_gc_duration_seconds{quantile=\"0\"} 3.3722e-05\n",
			expectedFormat: PromList{
				&model.Sample{
					Metric: model.Metric{
						model.MetricNameLabel: "go_gc_duration_seconds",
						"quantile":            "0",
						"instance":            "hostname",
						"prom_type":           "untyped",
					},
					Value:     3.3722e-05,
					Timestamp: model.TimeFromUnix(ts),
				},
			},
			outputMetricTags: []*v2.MetricTag{
				{
					Name:  "instance",
					Value: "hostname",
				},
			},
		},
		{
			metric: "go_gc_duration_seconds{quantile=\"0\"} 3.3722e-05\n",
			expectedFormat: PromList{
				&model.Sample{
					Metric: model.Metric{
						model.MetricNameLabel: "go_gc_duration_seconds",
						"quantile":            "0",
						"prom_type":           "untyped",
					},
					Value:     3.3722e-05,
					Timestamp: model.TimeFromUnix(ts),
				},
			},
		},
		{
			metric: "go_gc_duration_seconds{quantile=\"0\"} 3.3722e-05\n",
			expectedFormat: PromList{
				&model.Sample{
					Metric: model.Metric{
						model.MetricNameLabel: "go_gc_duration_seconds",
						"quantile":            "0",
						"prom_type":           "untyped",
					},
					Value:     3.3722e-05,
					Timestamp: model.TimeFromUnix(ts),
				},
			},
			outputMetricTags: []*v2.MetricTag{},
		},
		{
			metric: "go_gc_duration_seconds{quantile=\"0\"} 3.3722e-05\n",
			expectedFormat: PromList{
				&model.Sample{
					Metric: model.Metric{
						model.MetricNameLabel: "go_gc_duration_seconds",
						"quantile":            "0",
						"foo":                 "bar",
						"boo":                 "baz",
						"prom_type":           "untyped",
					},
					Value:     3.3722e-05,
					Timestamp: model.TimeFromUnix(ts),
				},
			},
			outputMetricTags: []*v2.MetricTag{
				{
					Name:  "foo",
					Value: "bar",
				},
				{
					Name:  "boo",
					Value: "baz",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			event := v2.FixtureEvent("test", "test")
			event.Check.Output = tc.metric
			event.Check.OutputMetricTags = tc.outputMetricTags
			prom := ParseProm(event)
			if !tc.timeInconclusive {
				assert.Equal(tc.expectedFormat, prom)
			}
		})
	}
}

func TestTransformProm(t *testing.T) {
	assert := assert.New(t)
	ts := time.Now().Unix()

	testCases := []struct {
		metric         PromList
		expectedFormat []*v2.MetricPoint
	}{
		{
			metric: PromList{
				&model.Sample{
					Metric: model.Metric{
						model.MetricNameLabel: "go_gc_duration_seconds",
						"quantile":            "0",
					},
					Value:     3.3722e-05,
					Timestamp: model.TimeFromUnix(ts),
				},
			},
			expectedFormat: []*v2.MetricPoint{
				{
					Name:      "go_gc_duration_seconds",
					Value:     3.3722e-05,
					Timestamp: ts,
					Tags: []*v2.MetricTag{
						{
							Name:  "quantile",
							Value: "0",
						},
					},
				},
			},
		},
		{
			metric: PromList{
				&model.Sample{
					Metric: model.Metric{
						model.MetricNameLabel: "go_memstats_alloc_bytes_total",
					},
					Value:     4.095146016e+09,
					Timestamp: model.TimeFromUnix(ts),
				},
			},
			expectedFormat: []*v2.MetricPoint{
				{
					Name:      "go_memstats_alloc_bytes_total",
					Value:     4.095146016e+09,
					Timestamp: ts,
					Tags:      []*v2.MetricTag{},
				},
			},
		},
		{
			metric: PromList{
				&model.Sample{
					Metric: model.Metric{
						model.MetricNameLabel: "go_memstats_alloc_bytes_total",
					},
					Value:     model.SampleValue(math.NaN()),
					Timestamp: model.TimeFromUnix(ts),
				},
			},
			expectedFormat: nil,
		},
	}

	for _, tc := range testCases {
		t.Run("transform", func(t *testing.T) {
			prom := tc.metric.Transform()
			assert.Equal(tc.expectedFormat, prom)
		})
	}
}
