package transformers

import (
	"reflect"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestParseNagios(t *testing.T) {
	testCases := []struct {
		name    string
		event   *types.Event
		want    NagiosList
		wantErr bool
	}{
		{
			name: "no perfdata metric",
			event: &types.Event{
				Check: &types.Check{
					Output: "PING ok - Packet loss = 0%, RTA = 0.80 ms",
				},
			},
			want:    NagiosList(nil),
			wantErr: true,
		},
		{
			name: "single perfdata metric",
			event: &types.Event{
				Check: &types.Check{
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
			wantErr: false,
		},
		{
			name: "multiple perfdata metrics",
			event: &types.Event{
				Check: &types.Check{
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
			wantErr: false,
		},
		{
			name: "invalid perfdata format",
			event: &types.Event{
				Check: &types.Check{
					Output: "PING ok - Packet loss = 0%, RTA = 0.80 ms | percent_packet_loss",
				},
			},
			want:    NagiosList(nil),
			wantErr: true,
		},
		{
			name: "invalid perfdata value",
			event: &types.Event{
				Check: &types.Check{
					Output: "PING ok - Packet loss = 0%, RTA = 0.80 ms | percent_packet_loss=a",
				},
			},
			want:    NagiosList(nil),
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseNagios(tc.event)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseNagios() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !assert.Equal(t, got, tc.want) {
				t.Errorf("ParseNagios() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTransformNagios(t *testing.T) {
	testCases := []struct {
		name    string
		metrics NagiosList
		want    []*types.MetricPoint
	}{
		{
			metrics: NagiosList{
				{
					Label:     "percent_packet_loss",
					Value:     0,
					Timestamp: 123456789,
				},
			},
			want: []*types.MetricPoint{
				{
					Name:      "percent_packet_loss",
					Value:     0,
					Timestamp: 123456789,
					Tags:      []*types.MetricTag{},
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
		name    string
		event   *types.Event
		want    []*types.MetricPoint
		wantErr bool
	}{
		{
			name: "happy path",
			event: &types.Event{
				Check: &types.Check{
					Executed: 123456789,
					Output:   "PING ok - Packet loss = 0% | percent_packet_loss=0",
				},
			},
			want: []*types.MetricPoint{
				{
					Name:      "percent_packet_loss",
					Value:     0,
					Timestamp: 123456789,
					Tags:      []*types.MetricTag{},
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			transformer, err := ParseNagios(tc.event)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseNagios() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			got := transformer.Transform()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParseNagios() = %v, want %v", got, tc.want)
			}
		})
	}
}
