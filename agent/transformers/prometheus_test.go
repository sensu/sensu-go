package transformers

import (
	"math"
	"testing"

	time "github.com/echlebek/timeproxy"

	"github.com/prometheus/common/model"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			event := types.FixtureEvent("test", "test")
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
		outputMetricTags []*types.MetricTag
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
			outputMetricTags: []*types.MetricTag{
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
			outputMetricTags: []*types.MetricTag{},
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
			outputMetricTags: []*types.MetricTag{
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
			event := types.FixtureEvent("test", "test")
			event.Check.Output = tc.metric
			event.Check.OutputMetricTags = tc.outputMetricTags
			prom := ParseProm(event)
			if !tc.timeInconclusive {
				assert.Equal(tc.expectedFormat, prom)
			}
		})
	}
}

func TestParsePromValidationScheme(t *testing.T) {
	assert := assert.New(t)

	event := types.FixtureEvent("test", "test")
	event.Check.Output = "# HELP node_cpu_seconds_total Seconds the CPUs spent in each mode.\n# TYPE node_cpu_seconds_total counter\nnode_cpu_seconds_total{cpu=\"0\",mode=\"idle\"} 123456.78\n"

	assert.NotPanics(func() {
		prom := ParseProm(event)
		require.NotEmpty(t, prom)
		points := prom.Transform()
		require.True(t, len(points) > 0)
		assert.Equal("node_cpu_seconds_total", points[0].Name)
	})
}

func TestParsePromCounter(t *testing.T) {
	event := types.FixtureEvent("test", "test")
	event.Check.Output = "# HELP http_requests_total The total number of HTTP requests.\n# TYPE http_requests_total counter\nhttp_requests_total{method=\"post\",code=\"200\"} 1027\nhttp_requests_total{method=\"post\",code=\"400\"} 3\n"

	prom := ParseProm(event)
	require.NotEmpty(t, prom)
	points := prom.Transform()
	require.Equal(t, 2, len(points))
	assert.Equal(t, "http_requests_total", points[0].Name)
	assert.Equal(t, "http_requests_total", points[1].Name)
}

func TestParsePromGauge(t *testing.T) {
	event := types.FixtureEvent("test", "test")
	event.Check.Output = "# HELP node_memory_MemAvailable_bytes Memory information field MemAvailable_bytes.\n# TYPE node_memory_MemAvailable_bytes gauge\nnode_memory_MemAvailable_bytes 5.361664e+09\n"

	prom := ParseProm(event)
	require.NotEmpty(t, prom)
	points := prom.Transform()
	require.Equal(t, 1, len(points))
	assert.Equal(t, "node_memory_MemAvailable_bytes", points[0].Name)
	assert.Equal(t, 5.361664e+09, points[0].Value)
}

func TestParsePromHistogram(t *testing.T) {
	assert := assert.New(t)

	event := types.FixtureEvent("test", "test")
	event.Check.Output = "# HELP http_request_duration_seconds A histogram of the request duration.\n# TYPE http_request_duration_seconds histogram\nhttp_request_duration_seconds_bucket{le=\"0.05\"} 24054\nhttp_request_duration_seconds_bucket{le=\"0.1\"} 33444\nhttp_request_duration_seconds_bucket{le=\"0.2\"} 100392\nhttp_request_duration_seconds_bucket{le=\"+Inf\"} 144320\nhttp_request_duration_seconds_sum 53423\nhttp_request_duration_seconds_count 144320\n"

	prom := ParseProm(event)
	assert.NotEmpty(prom)
	points := prom.Transform()
	assert.True(len(points) >= 6)

	nameSet := make(map[string]bool)
	for _, p := range points {
		nameSet[p.Name] = true
	}
	assert.True(nameSet["http_request_duration_seconds_bucket"])
	assert.True(nameSet["http_request_duration_seconds_sum"])
	assert.True(nameSet["http_request_duration_seconds_count"])
}

func TestParsePromSummary(t *testing.T) {
	assert := assert.New(t)

	event := types.FixtureEvent("test", "test")
	event.Check.Output = "# HELP go_gc_duration_seconds A summary of pause duration of garbage collection cycles.\n# TYPE go_gc_duration_seconds summary\ngo_gc_duration_seconds{quantile=\"0\"} 3.3722e-05\ngo_gc_duration_seconds{quantile=\"0.25\"} 5.0129e-05\ngo_gc_duration_seconds{quantile=\"0.5\"} 0.000126547\ngo_gc_duration_seconds{quantile=\"1\"} 0.009688629\ngo_gc_duration_seconds_sum 17.391350544\ngo_gc_duration_seconds_count 12345\n"

	prom := ParseProm(event)
	assert.NotEmpty(prom)
	points := prom.Transform()
	assert.True(len(points) >= 6)

	nameSet := make(map[string]bool)
	for _, p := range points {
		nameSet[p.Name] = true
	}
	assert.True(nameSet["go_gc_duration_seconds"])
	assert.True(nameSet["go_gc_duration_seconds_sum"])
	assert.True(nameSet["go_gc_duration_seconds_count"])
}

func TestParsePromMultipleMetricFamilies(t *testing.T) {
	assert := assert.New(t)

	event := types.FixtureEvent("test", "test")
	event.Check.Output = "# HELP node_filesystem_avail_bytes Filesystem space available.\n# TYPE node_filesystem_avail_bytes gauge\nnode_filesystem_avail_bytes{device=\"/dev/sda1\",mountpoint=\"/\"} 5.3687091e+10\n# HELP node_filesystem_size_bytes Filesystem size in bytes.\n# TYPE node_filesystem_size_bytes gauge\nnode_filesystem_size_bytes{device=\"/dev/sda1\",mountpoint=\"/\"} 1.03079215104e+11\n# HELP node_cpu_seconds_total Seconds the CPUs spent in each mode.\n# TYPE node_cpu_seconds_total counter\nnode_cpu_seconds_total{cpu=\"0\",mode=\"idle\"} 123456.78\nnode_cpu_seconds_total{cpu=\"0\",mode=\"system\"} 4567.89\n"

	prom := ParseProm(event)
	assert.NotEmpty(prom)
	points := prom.Transform()
	assert.Equal(4, len(points))

	nameSet := make(map[string]bool)
	for _, p := range points {
		nameSet[p.Name] = true
	}
	assert.True(nameSet["node_filesystem_avail_bytes"])
	assert.True(nameSet["node_filesystem_size_bytes"])
	assert.True(nameSet["node_cpu_seconds_total"])
}

func TestParsePromMalformedInput(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		name   string
		output string
	}{
		{"completely invalid", "this is not prometheus text at all"},
		{"partial metric line", "http_requests_total{method=\"get\"}\n"},
		{"missing value", "# TYPE foo counter\nfoo\n"},
		{"truncated output", "# HELP foo A help message.\n# TYPE foo counter\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := types.FixtureEvent("test", "test")
			event.Check.Output = tc.output

			assert.NotPanics(func() {
				_ = ParseProm(event)
			})
		})
	}
}

func TestParsePromEmptyOutput(t *testing.T) {
	assert := assert.New(t)

	event := types.FixtureEvent("test", "test")
	event.Check.Output = ""

	prom := ParseProm(event)
	assert.Empty(prom)
	points := prom.Transform()
	assert.Nil(points)
}

func TestParsePromNaN(t *testing.T) {
	nanList := PromList{
		&model.Sample{
			Metric: model.Metric{
				model.MetricNameLabel: "stale_metric",
			},
			Value:     model.SampleValue(math.NaN()),
			Timestamp: model.TimeFromUnix(time.Now().Unix()),
		},
		&model.Sample{
			Metric: model.Metric{
				model.MetricNameLabel: "valid_metric",
			},
			Value:     42.0,
			Timestamp: model.TimeFromUnix(time.Now().Unix()),
		},
	}

	points := nanList.Transform()
	require.Equal(t, 1, len(points))
	assert.Equal(t, "valid_metric", points[0].Name)
	assert.Equal(t, 42.0, points[0].Value)
}

func TestParsePromSpecialCharLabels(t *testing.T) {
	event := types.FixtureEvent("test", "test")
	event.Check.Output = "# TYPE http_requests_total counter\nhttp_requests_total{method=\"GET\",path=\"/api/v1/users\",status_code=\"200\"} 1500\n"

	prom := ParseProm(event)
	require.NotEmpty(t, prom)
	points := prom.Transform()
	require.Equal(t, 1, len(points))
	assert.Equal(t, "http_requests_total", points[0].Name)

	tagMap := make(map[string]string)
	for _, tag := range points[0].Tags {
		tagMap[tag.Name] = tag.Value
	}
	assert.Equal(t, "GET", tagMap["method"])
	assert.Equal(t, "/api/v1/users", tagMap["path"])
	assert.Equal(t, "200", tagMap["status_code"])
}

func TestTransformProm(t *testing.T) {
	assert := assert.New(t)
	ts := time.Now().Unix()

	testCases := []struct {
		metric         PromList
		expectedFormat []*types.MetricPoint
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
			expectedFormat: []*types.MetricPoint{
				{
					Name:      "go_gc_duration_seconds",
					Value:     3.3722e-05,
					Timestamp: ts,
					Tags: []*types.MetricTag{
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
			expectedFormat: []*types.MetricPoint{
				{
					Name:      "go_memstats_alloc_bytes_total",
					Value:     4.095146016e+09,
					Timestamp: ts,
					Tags:      []*types.MetricTag{},
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
