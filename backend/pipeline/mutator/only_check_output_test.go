package mutator

import (
	"context"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestOnlyCheckOutputAdapter_Name(t *testing.T) {
	o := &OnlyCheckOutputAdapter{}
	want := "OnlyCheckOutputAdapter"

	if got := o.Name(); want != got {
		t.Errorf("OnlyCheckOutputAdapter.Name() = %v, want %v", got, want)
	}
}

func TestOnlyCheckOutputAdapter_CanMutate(t *testing.T) {
	type args struct {
		ref *corev2.ResourceReference
	}
	tests := []struct {
		name string
		o    *OnlyCheckOutputAdapter
		args args
		want bool
	}{
		{
			name: "returns false when resource reference is not a core/v2.Mutator",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Handler",
				},
			},
			want: false,
		},
		{
			name: "returns false when resource reference is a core/v2.Mutator and its name is not only_check_output",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Mutator",
					Name:       "json",
				},
			},
			want: false,
		},
		{
			name: "returns true when resource reference is a core/v2.Mutator and its name is only_check_output",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Mutator",
					Name:       "only_check_output",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OnlyCheckOutputAdapter{}
			if got := o.CanMutate(tt.args.ref); got != tt.want {
				t.Errorf("OnlyCheckOutputAdapter.CanMutate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOnlyCheckOutputAdapter_Mutate(t *testing.T) {
	type args struct {
		ctx   context.Context
		ref   *corev2.ResourceReference
		event *corev2.Event
	}
	tests := []struct {
		name    string
		o       *OnlyCheckOutputAdapter
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "can mutate without producing any errors",
			args: args{
				event: &corev2.Event{
					Check: &corev2.Check{
						Output: "command executed successfully",
					},
				},
			},
			wantErr: false,
			want:    []byte("command executed successfully"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OnlyCheckOutputAdapter{}
			got, err := o.Mutate(tt.args.ctx, tt.args.ref, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("OnlyCheckOutputAdapter.Mutate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OnlyCheckOutputAdapter.Mutate() = %v, want %v", got, tt.want)
			}
		})
	}
}
