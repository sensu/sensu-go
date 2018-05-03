package transformers

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestParseInflux(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		metric         string
		expectedFormat InfluxList
		expectedErr    bool
	}{
		{
			metric: "weather,location=us-midwest,season=summer temperature=82,humidity=30 1465839830100400200",
			expectedFormat: InfluxList{
				{
					Measurement: "weather",
					TagSet: []*types.MetricTag{
						{
							Name:  "location",
							Value: "us-midwest",
						},
						{
							Name:  "season",
							Value: "summer",
						},
					},
					FieldSet: []*Field{
						{
							Key:   "temperature",
							Value: 82,
						},
						{
							Key:   "humidity",
							Value: 30,
						},
					},
					Timestamp: 1465839830,
				},
			},
			expectedErr: false,
		},
		{
			metric: "weather,location=us-midwest,season=summer temperature=82,humidity=30 1465839830100400200\nweather temperature=82 1465839830100400200",
			expectedFormat: InfluxList{
				{
					Measurement: "weather",
					TagSet: []*types.MetricTag{
						{
							Name:  "location",
							Value: "us-midwest",
						},
						{
							Name:  "season",
							Value: "summer",
						},
					},
					FieldSet: []*Field{
						{
							Key:   "temperature",
							Value: 82,
						},
						{
							Key:   "humidity",
							Value: 30,
						},
					},
					Timestamp: 1465839830,
				},
				{
					Measurement: "weather",
					TagSet:      []*types.MetricTag{},
					FieldSet: []*Field{
						{
							Key:   "temperature",
							Value: 82,
						},
					},
					Timestamp: 1465839830,
				},
			},
			expectedErr: false,
		},
		{
			metric: "weather temperature=82 1465839830100400200",
			expectedFormat: InfluxList{
				{
					Measurement: "weather",
					TagSet:      []*types.MetricTag{},
					FieldSet: []*Field{
						{
							Key:   "temperature",
							Value: 82,
						},
					},
					Timestamp: 1465839830,
				},
			},
			expectedErr: false,
		},
		{
			metric:         "weather temperature=82",
			expectedFormat: InfluxList{},
			expectedErr:    true,
		},
		{
			metric:         "weather,location temperature= 1465839830100400200",
			expectedFormat: InfluxList{},
			expectedErr:    true,
		},
		{
			metric:         "",
			expectedFormat: InfluxList{},
			expectedErr:    true,
		},
		{
			metric:         "foo bar baz",
			expectedFormat: InfluxList{},
			expectedErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			graphite, err := ParseInflux(tc.metric)
			if tc.expectedErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tc.expectedFormat, graphite)
		})
	}
}

func TestTransformInflux(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		metric         InfluxList
		expectedFormat []*types.MetricPoint
	}{
		{
			metric: InfluxList{
				{
					Measurement: "weather",
					TagSet: []*types.MetricTag{
						{
							Name:  "location",
							Value: "us-midwest",
						},
						{
							Name:  "season",
							Value: "summer",
						},
					},
					FieldSet: []*Field{
						{
							Key:   "temperature",
							Value: 82,
						},
						{
							Key:   "humidity",
							Value: 30,
						},
					},
					Timestamp: 1465839830,
				},
			},
			expectedFormat: []*types.MetricPoint{
				{
					Name:      "weather.temperature",
					Value:     82,
					Timestamp: 1465839830,
					Tags: []*types.MetricTag{
						{
							Name:  "location",
							Value: "us-midwest",
						},
						{
							Name:  "season",
							Value: "summer",
						},
					},
				},
				{
					Name:      "weather.humidity",
					Value:     30,
					Timestamp: 1465839830,
					Tags: []*types.MetricTag{
						{
							Name:  "location",
							Value: "us-midwest",
						},
						{
							Name:  "season",
							Value: "summer",
						},
					},
				},
			},
		},
		{
			metric: InfluxList{
				{
					Measurement: "",
					TagSet:      []*types.MetricTag{},
					FieldSet:    []*Field{},
					Timestamp:   0,
				},
			},
			expectedFormat: []*types.MetricPoint(nil),
		},
	}

	for _, tc := range testCases {
		t.Run("transform", func(t *testing.T) {
			graphite := tc.metric.Transform()
			assert.Equal(tc.expectedFormat, graphite)
		})
	}
}

func TestParseAndTransformInflux(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		metric         string
		expectedFormat []*types.MetricPoint
		expectedErr    bool
	}{
		{
			metric: "weather,location=us-midwest,season=summer temperature=82,humidity=30 1465839830100400200",
			expectedFormat: []*types.MetricPoint{
				{
					Name:      "weather.temperature",
					Value:     82,
					Timestamp: 1465839830,
					Tags: []*types.MetricTag{
						{
							Name:  "location",
							Value: "us-midwest",
						},
						{
							Name:  "season",
							Value: "summer",
						},
					},
				},
				{
					Name:      "weather.humidity",
					Value:     30,
					Timestamp: 1465839830,
					Tags: []*types.MetricTag{
						{
							Name:  "location",
							Value: "us-midwest",
						},
						{
							Name:  "season",
							Value: "summer",
						},
					},
				},
			},
			expectedErr: false,
		},
		{
			metric: "weather,location=us-midwest,season=summer temperature=82 1465839830100400200\nweather,location=us-midwest,season=summer humidity=30 1465839830100400200",
			expectedFormat: []*types.MetricPoint{
				{
					Name:      "weather.temperature",
					Value:     82,
					Timestamp: 1465839830,
					Tags: []*types.MetricTag{
						{
							Name:  "location",
							Value: "us-midwest",
						},
						{
							Name:  "season",
							Value: "summer",
						},
					},
				},
				{
					Name:      "weather.humidity",
					Value:     30,
					Timestamp: 1465839830,
					Tags: []*types.MetricTag{
						{
							Name:  "location",
							Value: "us-midwest",
						},
						{
							Name:  "season",
							Value: "summer",
						},
					},
				},
			},
			expectedErr: false,
		},
		{
			metric: "metric value=0 0\n",
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
			metric:      "weather temperature=82",
			expectedErr: true,
		},
		{
			metric:      "weather,location temperature= 1465839830100400200",
			expectedErr: true,
		},
		{
			metric:      "",
			expectedErr: true,
		},
		{
			metric:      "foo bar baz",
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			influx, err := ParseInflux(tc.metric)
			if tc.expectedErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				mp := influx.Transform()
				assert.Equal(tc.expectedFormat, mp)
			}
		})
	}
}
