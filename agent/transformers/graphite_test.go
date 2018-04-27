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
		expectedFormat Graphite
		expectedErr    bool
	}{
		{
			metric: "metric.value 1 123456789",
			expectedFormat: Graphite{
				Path:      "metric.value",
				Value:     1,
				Timestamp: 123456789,
			},
			expectedErr: false,
		},
		{
			metric:         "foo bar",
			expectedFormat: Graphite{},
			expectedErr:    true,
		},
		{
			metric:         "metric.value one 123456789",
			expectedFormat: Graphite{},
			expectedErr:    true,
		},
		{
			metric:         "metric.value 1 noon",
			expectedFormat: Graphite{},
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
		metric         Graphite
		expectedFormat []types.MetricPoint
	}{
		{
			metric: Graphite{
				Path:      "metric.value",
				Value:     1,
				Timestamp: 123456789,
			},
			expectedFormat: []types.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*types.MetricTag{},
				},
			},
		},
		{
			metric: Graphite{
				Path:      "",
				Value:     0,
				Timestamp: 0,
			},
			expectedFormat: []types.MetricPoint{
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
		expectedFormat []types.MetricPoint
		expectedErr    bool
	}{
		{
			metric: "metric.value 1 123456789",
			expectedFormat: []types.MetricPoint{
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
			metric: "metric.value 0 0",
			expectedFormat: []types.MetricPoint{
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
