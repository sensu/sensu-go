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
		t.Run(tc.metric, func(t *testing.T) {
			event := types.FixtureEvent("test", "test")
			event.Check.Output = tc.metric
			event.Check.OutputMetricTags = []*types.MetricTag{
				{
					Name:  "instance",
					Value: "hostname",
				},
			}
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
