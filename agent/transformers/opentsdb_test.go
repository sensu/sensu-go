package transformers

import (
	"reflect"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestParseOpenTSDB(t *testing.T) {
	testCases := []struct {
		name   string
		output string
		want   OpenTSDBList
	}{
		{
			name:   "standard opentsdb metric",
			output: "sys.cpu.user 1356998400 42.5 host=webserver01",
			want: OpenTSDBList{
				OpenTSDB{
					Name: "sys.cpu.user",
					TagSet: []*types.MetricTag{
						&types.MetricTag{
							Name:  "host",
							Value: "webserver01",
						},
					},
					Timestamp: 1356998400,
					Value:     42.5,
				},
			},
		},
		{
			name:   "standard opentsdb metric with whitespace",
			output: "sys.cpu.user 1356998400 42.5 host=webserver01 cpu=0\n",
			want: OpenTSDBList{
				OpenTSDB{
					Name: "sys.cpu.user",
					TagSet: []*types.MetricTag{
						&types.MetricTag{
							Name:  "host",
							Value: "webserver01",
						},
						&types.MetricTag{
							Name:  "cpu",
							Value: "0",
						},
					},
					Timestamp: 1356998400,
					Value:     42.5,
				},
			},
		},
		{
			name:   "GH_2511",
			output: "sys.cpu.user 1356998400 42.5 host=webserver01 cpu=0\nfoo",
			want: OpenTSDBList{
				OpenTSDB{
					Name: "sys.cpu.user",
					TagSet: []*types.MetricTag{
						&types.MetricTag{
							Name:  "host",
							Value: "webserver01",
						},
						&types.MetricTag{
							Name:  "cpu",
							Value: "0",
						},
					},
					Timestamp: 1356998400,
					Value:     42.5,
				},
			},
		},
		{
			name:   "timestamp with millisecond precision",
			output: "sys.cpu.user 1356998400000 42.5 host=webserver01",
			want: OpenTSDBList{
				OpenTSDB{
					Name: "sys.cpu.user",
					TagSet: []*types.MetricTag{
						&types.MetricTag{
							Name:  "host",
							Value: "webserver01",
						},
					},
					Timestamp: 1356998400,
					Value:     42.5,
				},
			},
		},
		{
			name:   "multiple tags",
			output: "sys.cpu.user 1356998400000 42.5 host=webserver01 cpu=0",
			want: OpenTSDBList{
				OpenTSDB{
					Name: "sys.cpu.user",
					TagSet: []*types.MetricTag{
						&types.MetricTag{
							Name:  "host",
							Value: "webserver01",
						},
						&types.MetricTag{
							Name:  "cpu",
							Value: "0",
						},
					},
					Timestamp: 1356998400,
					Value:     42.5,
				},
			},
		},
		{
			name:   "invalid format",
			output: "sys.cpu.user 1356998400000 42.5",
			want:   OpenTSDBList(nil),
		},
		{
			name:   "invalid timestamp",
			output: "sys.cpu.user foo 42.5 host=webserver01",
			want:   OpenTSDBList(nil),
		},
		{
			name:   "invalid value",
			output: "sys.cpu.user 1356998400 foo host=webserver01",
			want:   OpenTSDBList(nil),
		},
		{
			name:   "invalid tag",
			output: "sys.cpu.user 1356998400 42.5 host",
			want:   OpenTSDBList(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := types.FixtureEvent("test", "test")
			event.Check.Output = tc.output
			got := ParseOpenTSDB(event)
			if !assert.Equal(t, got, tc.want) {
				t.Errorf("ParseOpenTSDB() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseOpenTSDBTags(t *testing.T) {
	testCases := []struct {
		name   string
		output string
		want   OpenTSDBList
	}{
		{
			name:   "standard opentsdb metric",
			output: "sys.cpu.user 1356998400 42.5 host=webserver01",
			want: OpenTSDBList{
				OpenTSDB{
					Name: "sys.cpu.user",
					TagSet: []*types.MetricTag{
						&types.MetricTag{
							Name:  "host",
							Value: "webserver01",
						},
						&types.MetricTag{
							Name:  "instance",
							Value: "hostname",
						},
					},
					Timestamp: 1356998400,
					Value:     42.5,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := types.FixtureEvent("test", "test")
			event.Check.Output = tc.output
			event.Check.OutputMetricTags = []*types.MetricTag{
				{
					Name:  "instance",
					Value: "hostname",
				},
			}
			got := ParseOpenTSDB(event)
			if !assert.Equal(t, got, tc.want) {
				t.Errorf("ParseOpenTSDBTags() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTransformOpenTSDB(t *testing.T) {
	testCases := []struct {
		name    string
		metrics OpenTSDBList
		want    []*types.MetricPoint
	}{
		{
			metrics: OpenTSDBList{
				{
					Name: "sys.cpu.user",
					TagSet: []*types.MetricTag{
						&types.MetricTag{
							Name:  "host",
							Value: "webserver01",
						},
					},
					Timestamp: 1356998400,
					Value:     42.5,
				},
			},
			want: []*types.MetricPoint{
				{
					Name:      "sys.cpu.user",
					Value:     42.5,
					Timestamp: 1356998400,
					Tags: []*types.MetricTag{
						&types.MetricTag{
							Name:  "host",
							Value: "webserver01",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			points := tc.metrics.Transform()
			assert.Equal(t, tc.want, points)
		})
	}
}

func TestParseAndTransformOpenTSDB(t *testing.T) {
	testCases := []struct {
		name   string
		output string
		want   []*types.MetricPoint
	}{
		{
			name:   "happy path",
			output: "sys.cpu.user 1356998400 42.5 host=webserver01",
			want: []*types.MetricPoint{
				{
					Name:      "sys.cpu.user",
					Value:     42.5,
					Timestamp: 1356998400,
					Tags: []*types.MetricTag{
						&types.MetricTag{
							Name:  "host",
							Value: "webserver01",
						},
					},
				},
			},
		},
		{
			name:   "GH_2511",
			output: "sys.cpu.user 1356998400 42.5 host=webserver01\nfoo",
			want: []*types.MetricPoint{
				{
					Name:      "sys.cpu.user",
					Value:     42.5,
					Timestamp: 1356998400,
					Tags: []*types.MetricTag{
						&types.MetricTag{
							Name:  "host",
							Value: "webserver01",
						},
					},
				},
			},
		},
		{
			name:   "invalid metric",
			output: "sys.cpu.user 1356998400 42.5",
			want:   []*types.MetricPoint(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := types.FixtureEvent("test", "test")
			event.Check.Output = tc.output
			transformer := ParseOpenTSDB(event)
			got := transformer.Transform()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Transform() = %v, want %v", got, tc.want)
			}
		})
	}
}
