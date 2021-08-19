package mutator

import (
	"context"
	"reflect"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
)

func TestLegacy_Name(t *testing.T) {
	o := &Legacy{}
	want := "Legacy"

	if got := o.Name(); want != got {
		t.Errorf("Legacy.Name() = %v, want %v", got, want)
	}
}

func TestLegacy_CanMutate(t *testing.T) {
	type fields struct {
		AssetGetter            asset.Getter
		Executor               command.Executor
		SecretsProviderManager *secrets.ProviderManager
		Store                  store.Store
		StoreTimeout           time.Duration
	}
	type args struct {
		ctx context.Context
		ref *corev2.ResourceReference
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
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
			name: "returns false when resource reference is a core/v2.Mutator and its name is json",
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
			name: "returns false when resource reference is a core/v2.Mutator and its name is only_check_output",
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
			name: "returns true when resource reference is a core/v2.Mutator and its name doesn't match a built-in mutator",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Mutator",
					Name:       "my_unique_mutator",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Legacy{
				AssetGetter:            tt.fields.AssetGetter,
				Executor:               tt.fields.Executor,
				SecretsProviderManager: tt.fields.SecretsProviderManager,
				Store:                  tt.fields.Store,
				StoreTimeout:           tt.fields.StoreTimeout,
			}
			if got := l.CanMutate(tt.args.ctx, tt.args.ref); got != tt.want {
				t.Errorf("Legacy.CanMutate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLegacy_Mutate(t *testing.T) {
	type fields struct {
		AssetGetter            asset.Getter
		Executor               command.Executor
		SecretsProviderManager *secrets.ProviderManager
		Store                  store.Store
		StoreTimeout           time.Duration
	}
	type args struct {
		ctx   context.Context
		ref   *corev2.ResourceReference
		event *corev2.Event
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Legacy{
				AssetGetter:            tt.fields.AssetGetter,
				Executor:               tt.fields.Executor,
				SecretsProviderManager: tt.fields.SecretsProviderManager,
				Store:                  tt.fields.Store,
				StoreTimeout:           tt.fields.StoreTimeout,
			}
			got, err := l.Mutate(tt.args.ctx, tt.args.ref, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("Legacy.Mutate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Legacy.Mutate() = %v, want %v", got, tt.want)
			}
		})
	}
}
