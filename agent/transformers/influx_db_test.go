package transformers

import (
	"testing"

	v2 "github.com/sensu/core/v2"
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
					TagSet: []*v2.MetricTag{
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
					TagSet: []*v2.MetricTag{
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
					TagSet:      []*v2.MetricTag{},
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
					TagSet: []*v2.MetricTag{
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
					TagSet:      []*v2.MetricTag{},
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
			metric: "wea\\ ther,locat\\,ion=us-mid\\=west,sea\"son=sum\\\\mer te\\ mp\\,er\\=at\"ure=82,h\\ um\\,id\\=it\"y=30 1465839830100400200\nw\\ e\\,a\\=t\"her te\\ mp\\,er\\=at\"ure=82 1465839830100400200\n",
			expectedFormat: InfluxList{
				{
					Measurement: "wea ther",
					TagSet: []*v2.MetricTag{
						{
							Name:  "locat,ion",
							Value: "us-mid=west",
						},
						{
							Name:  `sea"son`,
							Value: `sum\mer`,
						},
					},
					FieldSet: []*Field{
						{
							Key:   `te mp,er=at"ure`,
							Value: 82,
						},
						{
							Key:   `h um,id=it"y`,
							Value: 30,
						},
					},
					Timestamp: 1465839830,
				},
				{
					Measurement: `w e,a=t"her`,
					TagSet:      []*v2.MetricTag{},
					FieldSet: []*Field{
						{
							Key:   `te mp,er=at"ure`,
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
			event := v2.FixtureEvent("test", "test")
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
		outputMetricTags []*v2.MetricTag
	}{
		{
			metric: "weather,location=us-midwest,season=summer temperature=82,humidity=30 1465839830100400200",
			expectedFormat: InfluxList{
				{
					Measurement: "weather",
					TagSet: []*v2.MetricTag{
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
			outputMetricTags: []*v2.MetricTag{
				{
					Name:  "instance",
					Value: "hostname",
				},
			},
		},
		{
			metric: "weather,location=us-midwest,season=summer temperature=82,humidity=30 1465839830100400200",
			expectedFormat: InfluxList{
				{
					Measurement: "weather",
					TagSet: []*v2.MetricTag{
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
			metric: "weather,location=us-midwest,season=summer temperature=82,humidity=30 1465839830100400200",
			expectedFormat: InfluxList{
				{
					Measurement: "weather",
					TagSet: []*v2.MetricTag{
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
			outputMetricTags: []*v2.MetricTag{},
		},
		{
			metric: "weather,location=us-midwest,season=summer temperature=82,humidity=30 1465839830100400200",
			expectedFormat: InfluxList{
				{
					Measurement: "weather",
					TagSet: []*v2.MetricTag{
						{
							Name:  "location",
							Value: "us-midwest",
						},
						{
							Name:  "season",
							Value: "summer",
						},
						{
							Name:  "foo",
							Value: "bar",
						},
						{
							Name:  "boo",
							Value: "baz",
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
			metric: "weather,location=us-midwest,season=summer temperature=82,humidity=30 1465839830100400200",
			expectedFormat: InfluxList{
				{
					Measurement: "weather",
					TagSet: []*v2.MetricTag{
						{
							Name:  "location",
							Value: "us-midwest",
						},
						{
							Name:  "season",
							Value: "summer",
						},
						{
							Name:  "",
							Value: "",
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
		expectedFormat []*v2.MetricPoint
	}{
		{
			metric: InfluxList{
				{
					Measurement: "weather",
					TagSet: []*v2.MetricTag{
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
			expectedFormat: []*v2.MetricPoint{
				{
					Name:      "weather.temperature",
					Value:     82,
					Timestamp: 1465839830,
					Tags: []*v2.MetricTag{
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
					Tags: []*v2.MetricTag{
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
					TagSet:      []*v2.MetricTag{},
					FieldSet:    []*Field{},
					Timestamp:   0,
				},
			},
			expectedFormat: []*v2.MetricPoint(nil),
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
		expectedFormat   []*v2.MetricPoint
		timeInconclusive bool
	}{
		{
			metric: "weather,location=us-midwest,season=summer temperature=82,humidity=30 1465839830100400200",
			expectedFormat: []*v2.MetricPoint{
				{
					Name:      "weather.temperature",
					Value:     82,
					Timestamp: 1465839830,
					Tags: []*v2.MetricTag{
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
					Tags: []*v2.MetricTag{
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
			expectedFormat: []*v2.MetricPoint{
				{
					Name:      "weather.temperature",
					Value:     82,
					Timestamp: 1465839830,
					Tags:      []*v2.MetricTag{},
				},
				{
					Name:      "weather.humidity",
					Value:     30,
					Timestamp: 1465839830,
					Tags:      []*v2.MetricTag{},
				},
			},
		},
		{
			metric: "weather,location=us-midwest,season=summer temperature=82 1465839830100400200\nweather,location=us-midwest,season=summer humidity=30 1465839830100400200",
			expectedFormat: []*v2.MetricPoint{
				{
					Name:      "weather.temperature",
					Value:     82,
					Timestamp: 1465839830,
					Tags: []*v2.MetricTag{
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
					Tags: []*v2.MetricTag{
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
			metric:           "weather temperature=82",
			timeInconclusive: true,
		},
		{
			metric:         "weather temperature=82 12345 blah",
			expectedFormat: []*v2.MetricPoint(nil),
		},
		{
			metric:         "weather,location temperature= 1465839830100400200",
			expectedFormat: []*v2.MetricPoint(nil),
		},
		{
			metric:         "",
			expectedFormat: []*v2.MetricPoint(nil),
		},
		{
			metric:         "foo bar baz",
			expectedFormat: []*v2.MetricPoint(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			event := v2.FixtureEvent("test", "test")
			event.Check.Output = tc.metric
			influx := ParseInflux(event)
			mp := influx.Transform()
			if !tc.timeInconclusive {
				assert.Equal(tc.expectedFormat, mp)
			}
		})
	}
}
