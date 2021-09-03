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

func TestPipeAdapter_Name(t *testing.T) {
	o := &PipeAdapter{}
	want := "PipeAdapter"

	if got := o.Name(); want != got {
		t.Errorf("PipeAdapter.Name() = %v, want %v", got, want)
	}
}

func TestPipeAdapter_CanMutate(t *testing.T) {
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
			name: "returns false",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PipeAdapter{
				AssetGetter:            tt.fields.AssetGetter,
				Executor:               tt.fields.Executor,
				SecretsProviderManager: tt.fields.SecretsProviderManager,
				Store:                  tt.fields.Store,
				StoreTimeout:           tt.fields.StoreTimeout,
			}
			if got := p.CanMutate(tt.args.ref); got != tt.want {
				t.Errorf("PipeAdapter.CanMutate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipeAdapter_Mutate(t *testing.T) {
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
		{
			name:    "returns error",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PipeAdapter{
				AssetGetter:            tt.fields.AssetGetter,
				Executor:               tt.fields.Executor,
				SecretsProviderManager: tt.fields.SecretsProviderManager,
				Store:                  tt.fields.Store,
				StoreTimeout:           tt.fields.StoreTimeout,
			}
			got, err := p.Mutate(tt.args.ctx, tt.args.ref, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("PipeAdapter.Mutate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PipeAdapter.Mutate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipeAdapter_run(t *testing.T) {
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
			p := &PipeAdapter{
				AssetGetter:            tt.fields.AssetGetter,
				Executor:               tt.fields.Executor,
				SecretsProviderManager: tt.fields.SecretsProviderManager,
				Store:                  tt.fields.Store,
				StoreTimeout:           tt.fields.StoreTimeout,
			}
			got, err := p.run(tt.args.ctx, tt.args.mutator, tt.args.event, tt.args.assets)
			if (err != nil) != tt.wantErr {
				t.Errorf("PipeAdapter.run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PipeAdapter.run() = %v, want %v", got, tt.want)
			}
		})
	}
}
