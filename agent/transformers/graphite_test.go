package transformers

import (
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
)

func TestParseGraphite(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		metric         string
		expectedFormat GraphiteList
	}{
		{
			metric: "metric.value 1 123456789",
			expectedFormat: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
				},
			},
		},
		{
			metric: "metric.value 1 123456789\n",
			expectedFormat: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
				},
			},
		},
		{
			metric: "metric.value 1 123456789\nmetric.value 0 0",
			expectedFormat: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
				},
				{
					Path:      "metric.value",
					Value:     0,
					Timestamp: 0,
				},
			},
		},
		{
			metric: "metric.value 1 123456789\nfoo",
			expectedFormat: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
				},
			},
		},
		{
			metric:         "",
			expectedFormat: GraphiteList(nil),
		},
		{
			metric:         "foo bar",
			expectedFormat: GraphiteList(nil),
		},
		{
			metric:         "metric.value one 123456789",
			expectedFormat: GraphiteList(nil),
		},
		{
			metric:         "metric.value 1 noon",
			expectedFormat: GraphiteList(nil),
		},
		{
			metric: "os.disk.used_bytes;GH=#4677;type=issue 2048 123456789",
			expectedFormat: GraphiteList{
				{
					Path:      "os.disk.used_bytes",
					Value:     2048,
					Timestamp: 123456789,
					Tags: []*v2.MetricTag{
						{Name: "GH", Value: "#4677"},
						{Name: "type", Value: "issue"},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			event := v2.FixtureEvent("test", "test")
			event.Check.Output = tc.metric
			graphite := ParseGraphite(event)
			assert.Equal(tc.expectedFormat, graphite)
		})
	}
}

func TestParseGraphiteTags(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		metric           string
		expectedFormat   GraphiteList
		outputMetricTags []*v2.MetricTag
	}{
		{
			metric: "metric.value 1 123456789",
			expectedFormat: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags: []*v2.MetricTag{
						&v2.MetricTag{
							Name:  "instance",
							Value: "hostname",
						},
					},
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
			metric: "metric.value 1 123456789",
			expectedFormat: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
				},
			},
		},
		{
			metric: "metric.value 1 123456789",
			expectedFormat: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*v2.MetricTag{},
				},
			},
			outputMetricTags: []*v2.MetricTag{},
		},
		{
			metric: "metric.value 1 123456789",
			expectedFormat: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags: []*v2.MetricTag{
						&v2.MetricTag{
							Name:  "foo",
							Value: "bar",
						},
						&v2.MetricTag{
							Name:  "boo",
							Value: "baz",
						},
					},
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
		{
			metric: "metric.value 1 123456789",
			expectedFormat: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags: []*v2.MetricTag{
						&v2.MetricTag{
							Name:  "",
							Value: "",
						},
					},
				},
			},
			outputMetricTags: []*v2.MetricTag{
				{
					Name:  "",
					Value: "",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			event := v2.FixtureEvent("test", "test")
			event.Check.Output = tc.metric
			event.Check.OutputMetricTags = tc.outputMetricTags
			graphite := ParseGraphite(event)
			assert.Equal(tc.expectedFormat, graphite)
		})
	}
}

func TestTransformGraphite(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		metric         GraphiteList
		expectedFormat []*v2.MetricPoint
	}{
		{
			metric: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
				},
			},
			expectedFormat: []*v2.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*v2.MetricTag{},
				},
			},
		},
		{
			metric: GraphiteList{
				{
					Path:      "",
					Value:     0,
					Timestamp: 0,
				},
			},
			expectedFormat: []*v2.MetricPoint{
				{
					Name:      "",
					Value:     0,
					Timestamp: 0,
					Tags:      []*v2.MetricTag{},
				},
			},
		},
		{
			metric: GraphiteList{
				{
					Path:      "",
					Value:     0,
					Timestamp: 0,
				},
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
				},
			},
			expectedFormat: []*v2.MetricPoint{
				{
					Name:      "",
					Value:     0,
					Timestamp: 0,
					Tags:      []*v2.MetricTag{},
				},
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*v2.MetricTag{},
				},
			},
		},
		{
			metric: GraphiteList{
				{
					Path:      "",
					Value:     0,
					Timestamp: 0,
				},
			},
			expectedFormat: []*v2.MetricPoint{
				{
					Name:      "",
					Value:     0,
					Timestamp: 0,
					Tags:      []*v2.MetricTag{},
				},
			},
		},
		{
			metric: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags: []*v2.MetricTag{
						&v2.MetricTag{
							Name:  "instance",
							Value: "hostname",
						},
					},
				},
			},
			expectedFormat: []*v2.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags: []*v2.MetricTag{
						&v2.MetricTag{
							Name:  "instance",
							Value: "hostname",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run("transform", func(t *testing.T) {
			graphite := tc.metric.Transform()
			assert.Equal(tc.expectedFormat, graphite)
		})
	}
}

func TestParseAndTransformGraphite(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		metric         string
		expectedFormat []*v2.MetricPoint
	}{
		{
			metric: "metric.value 1 123456789",
			expectedFormat: []*v2.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*v2.MetricTag{},
				},
			},
		},
		{
			metric: "metric.value 0 0\n",
			expectedFormat: []*v2.MetricPoint{
				{
					Name:      "metric.value",
					Value:     0,
					Timestamp: 0,
					Tags:      []*v2.MetricTag{},
				},
			},
		},
		{
			metric: "metric.value 1 123456789\nmetric.value 0 0",
			expectedFormat: []*v2.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*v2.MetricTag{},
				},
				{
					Name:      "metric.value",
					Value:     0,
					Timestamp: 0,
					Tags:      []*v2.MetricTag{},
				},
			},
		},
		{
			metric: "metric.value 1 123456789\nfoo",
			expectedFormat: []*v2.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*v2.MetricTag{},
				},
			},
		},
		{
			metric:         "",
			expectedFormat: []*v2.MetricPoint(nil),
		},
		{
			metric:         "foo bar",
			expectedFormat: []*v2.MetricPoint(nil),
		},
		{
			metric:         "metric.value one 123456789",
			expectedFormat: []*v2.MetricPoint(nil),
		},
		{
			metric:         "metric.value 1 noon",
			expectedFormat: []*v2.MetricPoint(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			event := v2.FixtureEvent("test", "test")
			event.Check.Output = tc.metric
			graphite := ParseGraphite(event)
			mp := graphite.Transform()
			assert.Equal(tc.expectedFormat, mp)
		})
	}
}
