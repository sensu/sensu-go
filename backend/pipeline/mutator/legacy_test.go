package mutator

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/js"
)

func TestHelperMutatorProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_MUTATOR_PROCESS") != "1" {
		return
	}

	command := strings.Join(os.Args[3:], " ")
	stdin, _ := ioutil.ReadAll(os.Stdin)

	switch command {
	case "cat":
		fmt.Fprintf(os.Stdout, "%s", stdin)
	}
	os.Exit(0)
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
				ctx: context.Background(),
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
				ctx: context.Background(),
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
				ctx: context.Background(),
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
				ctx: context.Background(),
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

func TestLegacy_pipeMutator(t *testing.T) {
	type fields struct {
		AssetGetter            asset.Getter
		Executor               command.Executor
		SecretsProviderManager *secrets.ProviderManager
		Store                  store.Store
		StoreTimeout           time.Duration
	}
	type args struct {
		ctx     context.Context
		mutator *corev2.Mutator
		event   *corev2.Event
		assets  asset.RuntimeAssetSet
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
			got, err := l.pipeMutator(tt.args.ctx, tt.args.mutator, tt.args.event, tt.args.assets)
			if (err != nil) != tt.wantErr {
				t.Errorf("Legacy.pipeMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Legacy.pipeMutator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLegacy_javascriptMutator(t *testing.T) {
	type fields struct {
		AssetGetter            asset.Getter
		Executor               command.Executor
		SecretsProviderManager *secrets.ProviderManager
		Store                  store.Store
		StoreTimeout           time.Duration
	}
	type args struct {
		ctx     context.Context
		mutator *corev2.Mutator
		event   *corev2.Event
		assets  js.JavascriptAssets
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
			got, err := l.javascriptMutator(tt.args.ctx, tt.args.mutator, tt.args.event, tt.args.assets)
			if (err != nil) != tt.wantErr {
				t.Errorf("Legacy.javascriptMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Legacy.javascriptMutator() = %v, want %v", got, tt.want)
			}
		})
	}
}
