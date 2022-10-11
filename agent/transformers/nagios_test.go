package transformers

import (
	"reflect"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
)

func TestParseNagios(t *testing.T) {
	testCases := []struct {
		name  string
		event *corev2.Event
		want  NagiosList
	}{
		{
			name: "no perfdata metric",
			event: &corev2.Event{
				Check: &corev2.Check{
					Output: "PING ok - Packet loss = 0%, RTA = 0.80 ms",
				},
			},
			want: NagiosList(nil),
		},
		{
			name: "single perfdata metric",
			event: &corev2.Event{
				Check: &corev2.Check{
					Executed: 12345,
					Output:   "PING ok - Packet loss = 0% | percent_packet_loss=0",
				},
			},
			want: NagiosList{
				Nagios{
					Label:     "percent_packet_loss",
					Value:     0.0,
					Timestamp: 12345,
				},
			},
		},
		{
			name: "single perfdata metric with newline",
			event: &corev2.Event{
				Check: &corev2.Check{
					Executed: 12345,
					Output:   "PING ok - Packet loss = 0% | percent_packet_loss=0\n",
				},
			},
			want: NagiosList{
				Nagios{
					Label:     "percent_packet_loss",
					Value:     0.0,
					Timestamp: 12345,
				},
			},
		},
		{
			name: "multiple perfdata metrics",
			event: &corev2.Event{
				Check: &corev2.Check{
					Executed: 12345,
					Output:   "PING ok - Packet loss = 0%, RTA = 0.80 ms | percent_packet_loss=0, rta=0.80",
				},
			},
			want: NagiosList{
				Nagios{
					Label:     "percent_packet_loss",
					Value:     0.0,
					Timestamp: 12345,
				},
				Nagios{
					Label:     "rta",
					Value:     0.8,
					Timestamp: 12345,
				},
			},
		},
		{
			name: "multiple perfdata metrics with output_metric_tags",
			event: &corev2.Event{
				Check: &corev2.Check{
					Executed: 12345,
					Output:   "PING ok - Packet loss = 0%, RTA = 0.80 ms | percent_packet_loss=0, rta=0.80",
					OutputMetricTags: []*corev2.MetricTag{
						{
							Name:  "foo",
							Value: "bar",
						},
					},
				},
			},
			want: NagiosList{
				Nagios{
					Label:     "percent_packet_loss",
					Value:     0.0,
					Timestamp: 12345,
					Tags: []*corev2.MetricTag{
						{
							Name:  "foo",
							Value: "bar",
						},
					},
				},
				Nagios{
					Label:     "rta",
					Value:     0.8,
					Timestamp: 12345,
					Tags: []*corev2.MetricTag{
						{
							Name:  "foo",
							Value: "bar",
						},
					},
				},
			},
		},
		{
			name: "multiple perfdata metrics with empty output_metric_tags",
			event: &corev2.Event{
				Check: &corev2.Check{
					Executed:         12345,
					Output:           "PING ok - Packet loss = 0%, RTA = 0.80 ms | percent_packet_loss=0, rta=0.80",
					OutputMetricTags: []*corev2.MetricTag{},
				},
			},
			want: NagiosList{
				Nagios{
					Label:     "percent_packet_loss",
					Value:     0.0,
					Timestamp: 12345,
					Tags:      []*corev2.MetricTag{},
				},
				Nagios{
					Label:     "rta",
					Value:     0.8,
					Timestamp: 12345,
					Tags:      []*corev2.MetricTag{},
				},
			},
		},
		{
			name: "multiple perfdata metrics with multiple output_metric_tags",
			event: &corev2.Event{
				Check: &corev2.Check{
					Executed: 12345,
					Output:   "PING ok - Packet loss = 0%, RTA = 0.80 ms | percent_packet_loss=0, rta=0.80",
					OutputMetricTags: []*corev2.MetricTag{
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
			},
			want: NagiosList{
				Nagios{
					Label:     "percent_packet_loss",
					Value:     0.0,
					Timestamp: 12345,
					Tags: []*corev2.MetricTag{
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
				Nagios{
					Label:     "rta",
					Value:     0.8,
					Timestamp: 12345,
					Tags: []*corev2.MetricTag{
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
			},
		},
		{
			name: "GH_2511",
			event: &corev2.Event{
				Check: &corev2.Check{
					Executed: 12345,
					Output:   "PING ok - Packet loss = 0%, RTA = 0.80 ms | percent_packet_loss=0, foo",
				},
			},
			want: NagiosList{
				Nagios{
					Label:     "percent_packet_loss",
					Value:     0.0,
					Timestamp: 12345,
				},
			},
		},
		{
			name: "invalid perfdata format",
			event: &corev2.Event{
				Check: &corev2.Check{
					Output: "PING ok - Packet loss = 0%, RTA = 0.80 ms | percent_packet_loss",
				},
			},
			want: NagiosList(nil),
		},
		{
			name: "invalid perfdata value",
			event: &corev2.Event{
				Check: &corev2.Check{
					Output: "PING ok - Packet loss = 0%, RTA = 0.80 ms | percent_packet_loss=a",
				},
			},
			want: NagiosList(nil),
		},
		{
			name: "bug #2021",
			event: &corev2.Event{
				Check: &corev2.Check{
					Executed: 12345,
					Output:   "CRITICAL - load average: 0.01, 0.04, 0.05|load1=0.010;0.010;0.010;0; load5=0.040;0.010;0.010;0; load15=0.050;0.010;0.010;0; \n",
				},
			},
			want: NagiosList{
				Nagios{
					Label:     "load1",
					Value:     0.01,
					Timestamp: 12345,
				},
				Nagios{
					Label:     "load5",
					Value:     0.04,
					Timestamp: 12345,
				},
				Nagios{
					Label:     "load15",
					Value:     0.05,
					Timestamp: 12345,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseNagios(tc.event)
			if !assert.Equal(t, tc.want, got) {
				t.Fatalf("ParseNagios() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTransformNagios(t *testing.T) {
	testCases := []struct {
		name    string
		metrics NagiosList
		want    []*corev2.MetricPoint
	}{
		{
			metrics: NagiosList{
				{
					Label:     "percent_packet_loss",
					Value:     0,
					Timestamp: 123456789,
				},
			},
			want: []*corev2.MetricPoint{
				{
					Name:      "percent_packet_loss",
					Value:     0,
					Timestamp: 123456789,
					Tags:      []*corev2.MetricTag{},
				},
			},
		},
		{
			metrics: NagiosList{
				{
					Label:     "percent_packet_loss",
					Value:     0,
					Timestamp: 123456789,
					Tags: []*corev2.MetricTag{
						{
							Name:  "foo",
							Value: "bar",
						},
					},
				},
			},
			want: []*corev2.MetricPoint{
				{
					Name:      "percent_packet_loss",
					Value:     0,
					Timestamp: 123456789,
					Tags: []*corev2.MetricTag{
						{
							Name:  "foo",
							Value: "bar",
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

func TestParseAndTransformNagios(t *testing.T) {
	testCases := []struct {
		name  string
		event *corev2.Event
		want  []*corev2.MetricPoint
	}{
		{
			name: "happy path",
			event: &corev2.Event{
				Check: &corev2.Check{
					Executed: 123456789,
					Output:   "PING ok - Packet loss = 0% | percent_packet_loss=0",
				},
			},
			want: []*corev2.MetricPoint{
				{
					Name:      "percent_packet_loss",
					Value:     0,
					Timestamp: 123456789,
					Tags:      []*corev2.MetricTag{},
				},
			},
		},
		{
			name: "GH_2511",
			event: &corev2.Event{
				Check: &corev2.Check{
					Executed: 123456789,
					Output:   "PING ok - Packet loss = 0% foo | percent_packet_loss=0 foo",
				},
			},
			want: []*corev2.MetricPoint{
				{
					Name:      "percent_packet_loss",
					Value:     0,
					Timestamp: 123456789,
					Tags:      []*corev2.MetricTag{},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			transformer := ParseNagios(tc.event)
			got := transformer.Transform()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParseNagios() = %v, want %v", got, tc.want)
			}
		})
	}
}
