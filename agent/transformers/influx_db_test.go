package transformers

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestParseInflux(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		metric           string
		expectedFormat   InfluxList
		timeInconclusive bool
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
		},
		{
			metric: "weather,location=us-midwest,season=summer temperature=82,humidity=30 1465839830100400200\nweather temperature=82 1465839830100400200\n",
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
		},
		{
			metric: "weather,location=us-midwest,season=summer temperature=82,humidity=30 1465839830100400200\nfoo\n",
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
		},
		{
			metric:           "weather temperature=82",
			timeInconclusive: true,
		},
		{
			metric:         "weather temperature=82 12345 blah",
			expectedFormat: InfluxList(nil),
		},
		{
			metric:         "weather,location temperature= 1465839830100400200",
			expectedFormat: InfluxList(nil),
		},
		{
			metric:         "",
			expectedFormat: InfluxList(nil),
		},
		{
			metric:         "foo bar baz",
			expectedFormat: InfluxList(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			event := types.FixtureEvent("test", "test")
			event.Check.Output = tc.metric
			influx := ParseInflux(event)
			if !tc.timeInconclusive {
				assert.Equal(tc.expectedFormat, influx)
			}
		})
	}
}

func TestParseInfluxTags(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		metric           string
		expectedFormat   InfluxList
		timeInconclusive bool
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
						{
							Name:  "instance",
							Value: "hostname",
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
			influx := ParseInflux(event)
			if !tc.timeInconclusive {
				assert.Equal(tc.expectedFormat, influx)
			}
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
		metric           string
		expectedFormat   []*types.MetricPoint
		timeInconclusive bool
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
		},
		{
			metric: "weather temperature=82,humidity=30 1465839830100400200\nfoo",
			expectedFormat: []*types.MetricPoint{
				{
					Name:      "weather.temperature",
					Value:     82,
					Timestamp: 1465839830,
					Tags:      []*types.MetricTag{},
				},
				{
					Name:      "weather.humidity",
					Value:     30,
					Timestamp: 1465839830,
					Tags:      []*types.MetricTag{},
				},
			},
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
		},
		{
			metric:           "weather temperature=82",
			timeInconclusive: true,
		},
		{
			metric:         "weather temperature=82 12345 blah",
			expectedFormat: []*types.MetricPoint(nil),
		},
		{
			metric:         "weather,location temperature= 1465839830100400200",
			expectedFormat: []*types.MetricPoint(nil),
		},
		{
			metric:         "",
			expectedFormat: []*types.MetricPoint(nil),
		},
		{
			metric:         "foo bar baz",
			expectedFormat: []*types.MetricPoint(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			event := types.FixtureEvent("test", "test")
			event.Check.Output = tc.metric
			influx := ParseInflux(event)
			mp := influx.Transform()
			if !tc.timeInconclusive {
				assert.Equal(tc.expectedFormat, mp)
			}
		})
	}
}
