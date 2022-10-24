package transformers

import (
	"testing"

	"github.com/sensu/sensu-go/types"
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
					Tags: []*types.MetricTag{
						{Name: "GH", Value: "#4677"},
						{Name: "type", Value: "issue"},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			event := types.FixtureEvent("test", "test")
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
		outputMetricTags []*types.MetricTag
	}{
		{
			metric: "metric.value 1 123456789",
			expectedFormat: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags: []*types.MetricTag{
						&types.MetricTag{
							Name:  "instance",
							Value: "hostname",
						},
					},
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
					Tags:      []*types.MetricTag{},
				},
			},
			outputMetricTags: []*types.MetricTag{},
		},
		{
			metric: "metric.value 1 123456789",
			expectedFormat: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags: []*types.MetricTag{
						&types.MetricTag{
							Name:  "foo",
							Value: "bar",
						},
						&types.MetricTag{
							Name:  "boo",
							Value: "baz",
						},
					},
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
		{
			metric: "metric.value 1 123456789",
			expectedFormat: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags: []*types.MetricTag{
						&types.MetricTag{
							Name:  "",
							Value: "",
						},
					},
				},
			},
			outputMetricTags: []*types.MetricTag{
				{
					Name:  "",
					Value: "",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			event := types.FixtureEvent("test", "test")
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
		expectedFormat []*types.MetricPoint
	}{
		{
			metric: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
				},
			},
			expectedFormat: []*types.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*types.MetricTag{},
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
			expectedFormat: []*types.MetricPoint{
				{
					Name:      "",
					Value:     0,
					Timestamp: 0,
					Tags:      []*types.MetricTag{},
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
			expectedFormat: []*types.MetricPoint{
				{
					Name:      "",
					Value:     0,
					Timestamp: 0,
					Tags:      []*types.MetricTag{},
				},
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*types.MetricTag{},
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
			expectedFormat: []*types.MetricPoint{
				{
					Name:      "",
					Value:     0,
					Timestamp: 0,
					Tags:      []*types.MetricTag{},
				},
			},
		},
		{
			metric: GraphiteList{
				{
					Path:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags: []*types.MetricTag{
						&types.MetricTag{
							Name:  "instance",
							Value: "hostname",
						},
					},
				},
			},
			expectedFormat: []*types.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags: []*types.MetricTag{
						&types.MetricTag{
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
		expectedFormat []*types.MetricPoint
	}{
		{
			metric: "metric.value 1 123456789",
			expectedFormat: []*types.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*types.MetricTag{},
				},
			},
		},
		{
			metric: "metric.value 0 0\n",
			expectedFormat: []*types.MetricPoint{
				{
					Name:      "metric.value",
					Value:     0,
					Timestamp: 0,
					Tags:      []*types.MetricTag{},
				},
			},
		},
		{
			metric: "metric.value 1 123456789\nmetric.value 0 0",
			expectedFormat: []*types.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*types.MetricTag{},
				},
				{
					Name:      "metric.value",
					Value:     0,
					Timestamp: 0,
					Tags:      []*types.MetricTag{},
				},
			},
		},
		{
			metric: "metric.value 1 123456789\nfoo",
			expectedFormat: []*types.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*types.MetricTag{},
				},
			},
		},
		{
			metric:         "",
			expectedFormat: []*types.MetricPoint(nil),
		},
		{
			metric:         "foo bar",
			expectedFormat: []*types.MetricPoint(nil),
		},
		{
			metric:         "metric.value one 123456789",
			expectedFormat: []*types.MetricPoint(nil),
		},
		{
			metric:         "metric.value 1 noon",
			expectedFormat: []*types.MetricPoint(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			event := types.FixtureEvent("test", "test")
			event.Check.Output = tc.metric
			graphite := ParseGraphite(event)
			mp := graphite.Transform()
			assert.Equal(tc.expectedFormat, mp)
		})
	}
}
