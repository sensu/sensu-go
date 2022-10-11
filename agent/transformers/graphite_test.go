package transformers

import (
	"testing"

	corev2 "github.com/sensu/core/v2"
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
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			event := corev2.FixtureEvent("test", "test")
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
		outputMetricTags []*corev2.MetricTag
	}{
		{
			metric: "metric.value 1 123456789",
			expectedFormat: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags: []*corev2.MetricTag{
						&corev2.MetricTag{
							Name:  "instance",
							Value: "hostname",
						},
					},
				},
			},
			outputMetricTags: []*corev2.MetricTag{
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
					Tags:      []*corev2.MetricTag{},
				},
			},
			outputMetricTags: []*corev2.MetricTag{},
		},
		{
			metric: "metric.value 1 123456789",
			expectedFormat: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags: []*corev2.MetricTag{
						&corev2.MetricTag{
							Name:  "foo",
							Value: "bar",
						},
						&corev2.MetricTag{
							Name:  "boo",
							Value: "baz",
						},
					},
				},
			},
			outputMetricTags: []*corev2.MetricTag{
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
					Tags: []*corev2.MetricTag{
						&corev2.MetricTag{
							Name:  "",
							Value: "",
						},
					},
				},
			},
			outputMetricTags: []*corev2.MetricTag{
				{
					Name:  "",
					Value: "",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			event := corev2.FixtureEvent("test", "test")
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
		expectedFormat []*corev2.MetricPoint
	}{
		{
			metric: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
				},
			},
			expectedFormat: []*corev2.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*corev2.MetricTag{},
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
			expectedFormat: []*corev2.MetricPoint{
				{
					Name:      "",
					Value:     0,
					Timestamp: 0,
					Tags:      []*corev2.MetricTag{},
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
			expectedFormat: []*corev2.MetricPoint{
				{
					Name:      "",
					Value:     0,
					Timestamp: 0,
					Tags:      []*corev2.MetricTag{},
				},
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*corev2.MetricTag{},
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
			expectedFormat: []*corev2.MetricPoint{
				{
					Name:      "",
					Value:     0,
					Timestamp: 0,
					Tags:      []*corev2.MetricTag{},
				},
			},
		},
		{
			metric: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags: []*corev2.MetricTag{
						&corev2.MetricTag{
							Name:  "instance",
							Value: "hostname",
						},
					},
				},
			},
			expectedFormat: []*corev2.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags: []*corev2.MetricTag{
						&corev2.MetricTag{
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
		expectedFormat []*corev2.MetricPoint
	}{
		{
			metric: "metric.value 1 123456789",
			expectedFormat: []*corev2.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*corev2.MetricTag{},
				},
			},
		},
		{
			metric: "metric.value 0 0\n",
			expectedFormat: []*corev2.MetricPoint{
				{
					Name:      "metric.value",
					Value:     0,
					Timestamp: 0,
					Tags:      []*corev2.MetricTag{},
				},
			},
		},
		{
			metric: "metric.value 1 123456789\nmetric.value 0 0",
			expectedFormat: []*corev2.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*corev2.MetricTag{},
				},
				{
					Name:      "metric.value",
					Value:     0,
					Timestamp: 0,
					Tags:      []*corev2.MetricTag{},
				},
			},
		},
		{
			metric: "metric.value 1 123456789\nfoo",
			expectedFormat: []*corev2.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*corev2.MetricTag{},
				},
			},
		},
		{
			metric:         "",
			expectedFormat: []*corev2.MetricPoint(nil),
		},
		{
			metric:         "foo bar",
			expectedFormat: []*corev2.MetricPoint(nil),
		},
		{
			metric:         "metric.value one 123456789",
			expectedFormat: []*corev2.MetricPoint(nil),
		},
		{
			metric:         "metric.value 1 noon",
			expectedFormat: []*corev2.MetricPoint(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			event := corev2.FixtureEvent("test", "test")
			event.Check.Output = tc.metric
			graphite := ParseGraphite(event)
			mp := graphite.Transform()
			assert.Equal(tc.expectedFormat, mp)
		})
	}
}
