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
		expectedErr    bool
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
			expectedErr: false,
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
			expectedErr: false,
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
			expectedErr: false,
		},
		{
			metric:         "",
			expectedFormat: GraphiteList{},
			expectedErr:    true,
		},
		{
			metric:         "foo bar",
			expectedFormat: GraphiteList{},
			expectedErr:    true,
		},
		{
			metric:         "metric.value one 123456789",
			expectedFormat: GraphiteList{},
			expectedErr:    true,
		},
		{
			metric:         "metric.value 1 noon",
			expectedFormat: GraphiteList{},
			expectedErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			graphite, err := ParseGraphite(tc.metric)
			if tc.expectedErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
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
		expectedErr    bool
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
			expectedErr: false,
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
			expectedErr: false,
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
			expectedErr: false,
		},
		{
			metric:      "",
			expectedErr: true,
		},
		{
			metric:      "foo bar",
			expectedErr: true,
		},
		{
			metric:      "metric.value one 123456789",
			expectedErr: true,
		},
		{
			metric:      "metric.value 1 noon",
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			graphite, err := ParseGraphite(tc.metric)
			if tc.expectedErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				mp := graphite.Transform()
				assert.Equal(tc.expectedFormat, mp)
			}
		})
	}
}
