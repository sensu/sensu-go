package mutator

import (
	"context"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestOnlyCheckOutput_Name(t *testing.T) {
	o := &OnlyCheckOutput{}
	want := "OnlyCheckOutput"

	if got := o.Name(); want != got {
		t.Errorf("OnlyCheckOutput.Name() = %v, want %v", got, want)
	}
}

func TestOnlyCheckOutput_CanMutate(t *testing.T) {
	type args struct {
		ctx context.Context
		ref *corev2.ResourceReference
	}
	tests := []struct {
		name string
		o    *OnlyCheckOutput
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
			o := &OnlyCheckOutput{}
			if got := o.CanMutate(tt.args.ctx, tt.args.ref); got != tt.want {
				t.Errorf("OnlyCheckOutput.CanMutate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOnlyCheckOutput_Mutate(t *testing.T) {
	type args struct {
		ctx   context.Context
		ref   *corev2.ResourceReference
		event *corev2.Event
	}
	tests := []struct {
		name    string
		o       *OnlyCheckOutput
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
			o := &OnlyCheckOutput{}
			got, err := o.Mutate(tt.args.ctx, tt.args.ref, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("OnlyCheckOutput.Mutate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OnlyCheckOutput.Mutate() = %v, want %v", got, tt.want)
			}
		})
	}
}
