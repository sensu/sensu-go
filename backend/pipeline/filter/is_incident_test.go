package filter

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestIsIncidentAdapter_Name(t *testing.T) {
	o := &IsIncidentAdapter{}
	want := "IsIncidentAdapter"

	if got := o.Name(); want != got {
		t.Errorf("IsIncidentAdapter.Name() = %v, want %v", got, want)
	}
}

func TestIsIncidentAdapter_CanFilter(t *testing.T) {
	type args struct {
		ctx context.Context
		ref *corev2.ResourceReference
	}
	tests := []struct {
		name string
		i    *IsIncidentAdapter
		args args
		want bool
	}{
		{
			name: "returns false when resource reference is not a core/v2.EventFilter",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Handler",
				},
			},
			want: false,
		},
		{
			name: "returns false when resource reference is a core/v2.EventFilter and its name is not is_incident",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "EventFilter",
					Name:       "has_metrics",
				},
			},
			want: false,
		},
		{
			name: "returns true when resource reference is a core/v2.EventFilter and its name is is_incident",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "EventFilter",
					Name:       "is_incident",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &IsIncidentAdapter{}
			if got := i.CanFilter(tt.args.ctx, tt.args.ref); got != tt.want {
				t.Errorf("IsIncidentAdapter.CanFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsIncidentAdapter_Filter(t *testing.T) {
	type args struct {
		ctx   context.Context
		ref   *corev2.ResourceReference
		event *corev2.Event
	}
	tests := []struct {
		name    string
		i       *IsIncidentAdapter
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "event is denied when it is not an incident or a resolution",
			args: args{
				ctx: context.Background(),
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("default", "default")
					event.Check.Status = 0
					return event
				}(),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "event is allowed when it is an incident",
			args: args{
				ctx: context.Background(),
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("default", "default")
					event.Check.Status = 1
					return event
				}(),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "event is allowed when it is a resolution",
			args: args{
				ctx: context.Background(),
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("default", "default")
					event.Check.Status = 0
					event.Check.History[len(event.Check.History)-2].Status = 1
					return event
				}(),
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &IsIncidentAdapter{}
			got, err := i.Filter(tt.args.ctx, tt.args.ref, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsIncidentAdapter.Filter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsIncidentAdapter.Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
