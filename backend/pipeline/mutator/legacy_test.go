package mutator

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLegacyAdapter_Name(t *testing.T) {
	o := &LegacyAdapter{}
	want := "LegacyAdapter"

	if got := o.Name(); want != got {
		t.Errorf("LegacyAdapter.Name() = %v, want %v", got, want)
	}
}

func TestLegacyAdapter_CanMutate(t *testing.T) {
	type fields struct {
		AssetGetter            asset.Getter
		Executor               command.Executor
		SecretsProviderManager *secrets.ProviderManager
		Store                  store.Store
		StoreTimeout           time.Duration
	}
	type args struct {
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
			l := &LegacyAdapter{
				AssetGetter:            tt.fields.AssetGetter,
				Executor:               tt.fields.Executor,
				SecretsProviderManager: tt.fields.SecretsProviderManager,
				Store:                  tt.fields.Store,
				StoreTimeout:           tt.fields.StoreTimeout,
			}
			if got := l.CanMutate(tt.args.ref); got != tt.want {
				t.Errorf("LegacyAdapter.CanMutate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLegacyAdapter_Mutate(t *testing.T) {
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
		wantFn  func(*corev2.Event) []byte
		wantErr bool
	}{
		{
			name: "can mutate using a pipe mutator",
			fields: fields{
				Store: func() store.Store {
					mutator := corev2.FakeMutatorCommand("cat")
					stor := &mockstore.MockStore{}
					stor.On("GetMutatorByName", mock.Anything, mock.Anything).Return(mutator, nil)
					return stor
				}(),
				Executor: command.NewExecutor(),
			},
			args: args{
				ctx: context.Background(),
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Mutator",
					Name:       "cat",
				},
				event: corev2.FixtureEvent("default", "default"),
			},
			wantFn: func(event *corev2.Event) []byte {
				bytes, _ := json.Marshal(event)
				return bytes
			},
			wantErr: false,
		},
		{
			name: "returns an error when a mutator cannot be found in the store (GH2784)",
			fields: fields{
				Store: func() store.Store {
					stor := &mockstore.MockStore{}
					stor.On("GetMutatorByName", mock.Anything, mock.Anything).Return((*corev2.Mutator)(nil), nil)
					return stor
				}(),
				Executor: command.NewExecutor(),
			},
			args: args{
				ctx: context.Background(),
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Mutator",
					Name:       "cat",
				},
				event: corev2.FixtureEvent("default", "default"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &LegacyAdapter{
				AssetGetter:            tt.fields.AssetGetter,
				Executor:               tt.fields.Executor,
				SecretsProviderManager: tt.fields.SecretsProviderManager,
				Store:                  tt.fields.Store,
				StoreTimeout:           tt.fields.StoreTimeout,
			}
			got, err := l.Mutate(tt.args.ctx, tt.args.ref, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("LegacyAdapter.Mutate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantFn != nil {
				want := tt.wantFn(tt.args.event)
				assert.JSONEq(t, string(want), string(got))
			}
		})
	}
}
