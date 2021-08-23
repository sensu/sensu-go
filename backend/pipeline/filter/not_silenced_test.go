package filter

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestNotSilenced_Name(t *testing.T) {
	o := &NotSilenced{}
	want := "NotSilenced"

	if got := o.Name(); want != got {
		t.Errorf("NotSilenced.Name() = %v, want %v", got, want)
	}
}

func TestNotSilenced_CanFilter(t *testing.T) {
	type args struct {
		ctx context.Context
		ref *corev2.ResourceReference
	}
	tests := []struct {
		name string
		i    *NotSilenced
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
			name: "returns false when resource reference is a core/v2.EventFilter and its name is not not_silenced",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "EventFilter",
					Name:       "is_incident",
				},
			},
			want: false,
		},
		{
			name: "returns true when resource reference is a core/v2.EventFilter and its name is not_silenced",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "EventFilter",
					Name:       "not_silenced",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &NotSilenced{}
			if got := i.CanFilter(tt.args.ctx, tt.args.ref); got != tt.want {
				t.Errorf("NotSilenced.CanFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotSilenced_Filter(t *testing.T) {
	type args struct {
		ctx   context.Context
		ref   *corev2.ResourceReference
		event *corev2.Event
	}
	tests := []struct {
		name    string
		i       *NotSilenced
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "event is denied when it is silenced",
			args: args{
				ctx: context.Background(),
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("default", "default")
					event.Check.Silenced = []string{"silenced_entry"}
					return event
				}(),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "event is allowed when it not silenced",
			args: args{
				ctx: context.Background(),
				event: func() *corev2.Event {
					event := corev2.FixtureEvent("default", "default")
					return event
				}(),
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &NotSilenced{}
			got, err := i.Filter(tt.args.ctx, tt.args.ref, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("NotSilenced.Filter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NotSilenced.Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
