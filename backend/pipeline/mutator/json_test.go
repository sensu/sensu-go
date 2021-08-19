package mutator

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestJSON_Name(t *testing.T) {
	o := &JSON{}
	want := "JSON"

	if got := o.Name(); want != got {
		t.Errorf("JSON.Name() = %v, want %v", got, want)
	}
}

func TestJSON_CanMutate(t *testing.T) {
	type args struct {
		ctx context.Context
		ref *corev2.ResourceReference
	}
	tests := []struct {
		name string
		j    *JSON
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
			name: "returns false when resource reference is a core/v2.Mutator and its name is not json",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Mutator",
					Name:       "only_check_output",
				},
			},
			want: false,
		},
		{
			name: "returns true when resource reference is a core/v2.Mutator and its name is json",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Mutator",
					Name:       "json",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JSON{}
			if got := j.CanMutate(tt.args.ctx, tt.args.ref); got != tt.want {
				t.Errorf("JSON.CanMutate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSON_Mutate(t *testing.T) {
	type args struct {
		ctx   context.Context
		ref   *corev2.ResourceReference
		event *corev2.Event
	}
	tests := []struct {
		name    string
		j       *JSON
		args    args
		wantFn  func(*corev2.Event) []byte
		wantErr bool
	}{
		{
			name: "can mutate without producing any errors",
			args: args{
				event: &corev2.Event{},
			},
			wantErr: false,
			wantFn: func(event *corev2.Event) []byte {
				bytes, _ := json.Marshal(event)
				return bytes
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JSON{}
			got, err := j.Mutate(tt.args.ctx, tt.args.ref, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSON.Mutate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := tt.wantFn(tt.args.event)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("JSON.Mutate() = %v, want %v", got, want)
			}
		})
	}
}
