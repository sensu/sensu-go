package transformers

import (
	"reflect"
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
)

func TestParseOpenTSDB(t *testing.T) {
	testCases := []struct {
		name	string
		output	string
		want	OpenTSDBList
	}{
		{
			name:	"standard opentsdb metric",
			output:	"sys.cpu.user 1356998400 42.5 host=webserver01",
			want: OpenTSDBList{
				OpenTSDB{
					Name:	"sys.cpu.user",
					TagSet: []*v2.MetricTag{
						&v2.MetricTag{
							Name:	"host",
							Value:	"webserver01",
						},
					},
					Timestamp:	1356998400,
					Value:		42.5,
				},
			},
		},
		{
			name:	"standard opentsdb metric with whitespace",
			output:	"sys.cpu.user 1356998400 42.5 host=webserver01 cpu=0\n",
			want: OpenTSDBList{
				OpenTSDB{
					Name:	"sys.cpu.user",
					TagSet: []*v2.MetricTag{
						&v2.MetricTag{
							Name:	"host",
							Value:	"webserver01",
						},
						&v2.MetricTag{
							Name:	"cpu",
							Value:	"0",
						},
					},
					Timestamp:	1356998400,
					Value:		42.5,
				},
			},
		},
		{
			name:	"GH_2511",
			output:	"sys.cpu.user 1356998400 42.5 host=webserver01 cpu=0\nfoo",
			want: OpenTSDBList{
				OpenTSDB{
					Name:	"sys.cpu.user",
					TagSet: []*v2.MetricTag{
						&v2.MetricTag{
							Name:	"host",
							Value:	"webserver01",
						},
						&v2.MetricTag{
							Name:	"cpu",
							Value:	"0",
						},
					},
					Timestamp:	1356998400,
					Value:		42.5,
				},
			},
		},
		{
			name:	"timestamp with millisecond precision",
			output:	"sys.cpu.user 1356998400000 42.5 host=webserver01",
			want: OpenTSDBList{
				OpenTSDB{
					Name:	"sys.cpu.user",
					TagSet: []*v2.MetricTag{
						&v2.MetricTag{
							Name:	"host",
							Value:	"webserver01",
						},
					},
					Timestamp:	1356998400,
					Value:		42.5,
				},
			},
		},
		{
			name:	"multiple tags",
			output:	"sys.cpu.user 1356998400000 42.5 host=webserver01 cpu=0",
			want: OpenTSDBList{
				OpenTSDB{
					Name:	"sys.cpu.user",
					TagSet: []*v2.MetricTag{
						&v2.MetricTag{
							Name:	"host",
							Value:	"webserver01",
						},
						&v2.MetricTag{
							Name:	"cpu",
							Value:	"0",
						},
					},
					Timestamp:	1356998400,
					Value:		42.5,
				},
			},
		},
		{
			name:	"invalid format",
			output:	"sys.cpu.user 1356998400000 42.5",
			want:	OpenTSDBList(nil),
		},
		{
			name:	"invalid timestamp",
			output:	"sys.cpu.user foo 42.5 host=webserver01",
			want:	OpenTSDBList(nil),
		},
		{
			name:	"invalid value",
			output:	"sys.cpu.user 1356998400 foo host=webserver01",
			want:	OpenTSDBList(nil),
		},
		{
			name:	"invalid tag",
			output:	"sys.cpu.user 1356998400 42.5 host",
			want:	OpenTSDBList(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := v2.FixtureEvent("test", "test")
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
		name			string
		output			string
		want			OpenTSDBList
		outputMetricTags	[]*v2.MetricTag
	}{
		{
			name:	"standard opentsdb metric with output metric tags",
			output:	"sys.cpu.user 1356998400 42.5 host=webserver01",
			want: OpenTSDBList{
				OpenTSDB{
					Name:	"sys.cpu.user",
					TagSet: []*v2.MetricTag{
						&v2.MetricTag{
							Name:	"host",
							Value:	"webserver01",
						},
						&v2.MetricTag{
							Name:	"instance",
							Value:	"hostname",
						},
					},
					Timestamp:	1356998400,
					Value:		42.5,
				},
			},
			outputMetricTags: []*v2.MetricTag{
				{
					Name:	"instance",
					Value:	"hostname",
				},
			},
		},
		{
			name:	"standard opentsdb metric with no output metric tags",
			output:	"sys.cpu.user 1356998400 42.5 host=webserver01",
			want: OpenTSDBList{
				OpenTSDB{
					Name:	"sys.cpu.user",
					TagSet: []*v2.MetricTag{
						&v2.MetricTag{
							Name:	"host",
							Value:	"webserver01",
						},
					},
					Timestamp:	1356998400,
					Value:		42.5,
				},
			},
		},
		{
			name:	"standard opentsdb metric with empty output metric tags",
			output:	"sys.cpu.user 1356998400 42.5 host=webserver01",
			want: OpenTSDBList{
				OpenTSDB{
					Name:	"sys.cpu.user",
					TagSet: []*v2.MetricTag{
						&v2.MetricTag{
							Name:	"host",
							Value:	"webserver01",
						},
					},
					Timestamp:	1356998400,
					Value:		42.5,
				},
			},
			outputMetricTags:	[]*v2.MetricTag{},
		},
		{
			name:	"standard opentsdb metric with multiple output metric tags",
			output:	"sys.cpu.user 1356998400 42.5 host=webserver01",
			want: OpenTSDBList{
				OpenTSDB{
					Name:	"sys.cpu.user",
					TagSet: []*v2.MetricTag{
						&v2.MetricTag{
							Name:	"host",
							Value:	"webserver01",
						},
						&v2.MetricTag{
							Name:	"foo",
							Value:	"bar",
						},
						&v2.MetricTag{
							Name:	"boo",
							Value:	"baz",
						},
					},
					Timestamp:	1356998400,
					Value:		42.5,
				},
			},
			outputMetricTags: []*v2.MetricTag{
				&v2.MetricTag{
					Name:	"foo",
					Value:	"bar",
				},
				&v2.MetricTag{
					Name:	"boo",
					Value:	"baz",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := v2.FixtureEvent("test", "test")
			event.Check.Output = tc.output
			event.Check.OutputMetricTags = tc.outputMetricTags
			got := ParseOpenTSDB(event)
			if !assert.Equal(t, got, tc.want) {
				t.Errorf("ParseOpenTSDBTags() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTransformOpenTSDB(t *testing.T) {
	testCases := []struct {
		name	string
		metrics	OpenTSDBList
		want	[]*v2.MetricPoint
	}{
		{
			metrics: OpenTSDBList{
				{
					Name:	"sys.cpu.user",
					TagSet: []*v2.MetricTag{
						&v2.MetricTag{
							Name:	"host",
							Value:	"webserver01",
						},
					},
					Timestamp:	1356998400,
					Value:		42.5,
				},
			},
			want: []*v2.MetricPoint{
				{
					Name:		"sys.cpu.user",
					Value:		42.5,
					Timestamp:	1356998400,
					Tags: []*v2.MetricTag{
						&v2.MetricTag{
							Name:	"host",
							Value:	"webserver01",
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
		name	string
		output	string
		want	[]*v2.MetricPoint
	}{
		{
			name:	"happy path",
			output:	"sys.cpu.user 1356998400 42.5 host=webserver01",
			want: []*v2.MetricPoint{
				{
					Name:		"sys.cpu.user",
					Value:		42.5,
					Timestamp:	1356998400,
					Tags: []*v2.MetricTag{
						&v2.MetricTag{
							Name:	"host",
							Value:	"webserver01",
						},
					},
				},
			},
		},
		{
			name:	"GH_2511",
			output:	"sys.cpu.user 1356998400 42.5 host=webserver01\nfoo",
			want: []*v2.MetricPoint{
				{
					Name:		"sys.cpu.user",
					Value:		42.5,
					Timestamp:	1356998400,
					Tags: []*v2.MetricTag{
						&v2.MetricTag{
							Name:	"host",
							Value:	"webserver01",
						},
					},
				},
			},
		},
		{
			name:	"invalid metric",
			output:	"sys.cpu.user 1356998400 42.5",
			want:	[]*v2.MetricPoint(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := v2.FixtureEvent("test", "test")
			event.Check.Output = tc.output
			transformer := ParseOpenTSDB(event)
			got := transformer.Transform()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Transform() = %v, want %v", got, tc.want)
			}
		})
	}
}
