package transformers

import (
	"reflect"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestParseOpenTSDB(t *testing.T) {
	testCases := []struct {
		name    string
		output  string
		want    OpenTSDBList
		wantErr bool
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
			wantErr: false,
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
			wantErr: false,
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
			wantErr: false,
		},
		{
			name:    "invalid format",
			output:  "sys.cpu.user 1356998400000 42.5",
			want:    OpenTSDBList(nil),
			wantErr: true,
		},
		{
			name:    "invalid timestamp",
			output:  "sys.cpu.user foo 42.5 host=webserver01",
			want:    OpenTSDBList(nil),
			wantErr: true,
		},
		{
			name:    "invalid value",
			output:  "sys.cpu.user 1356998400 foo host=webserver01",
			want:    OpenTSDBList(nil),
			wantErr: true,
		},
		{
			name:    "invalid tag",
			output:  "sys.cpu.user 1356998400 42.5 host",
			want:    OpenTSDBList(nil),
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseOpenTSDB(tc.output)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseOpenTSDB() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !assert.Equal(t, got, tc.want) {
				t.Errorf("ParseOpenTSDB() = %v, want %v", got, tc.want)
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
		name    string
		output  string
		want    []*types.MetricPoint
		wantErr bool
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
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			transformer, err := ParseOpenTSDB(tc.output)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseOpenTSDB() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			got := transformer.Transform()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Transform() = %v, want %v", got, tc.want)
			}
		})
	}
}
