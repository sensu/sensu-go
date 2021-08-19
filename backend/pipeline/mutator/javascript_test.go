package mutator

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/js"
	"github.com/stretchr/testify/assert"
)

type mutatorAssetSet struct{}

func (mutatorAssetSet) Key() string {
	return "mutatorAsset"
}

func (mutatorAssetSet) Scripts() (map[string]io.ReadCloser, error) {
	result := make(map[string]io.ReadCloser)
	result["mutatorAsset"] = ioutil.NopCloser(strings.NewReader(`var assetFunc = function () { event.check.labels["hockey"] = hockey; }`))
	return result, nil
}

func TestJavascript_Name(t *testing.T) {
	o := &Javascript{}
	want := "Javascript"

	if got := o.Name(); want != got {
		t.Errorf("Javascript.Name() = %v, want %v", got, want)
	}
}

func TestJavascript_CanMutate(t *testing.T) {
	type fields struct {
		AssetGetter  asset.Getter
		Store        store.Store
		StoreTimeout time.Duration
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
			name: "returns false",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &Javascript{
				AssetGetter:  tt.fields.AssetGetter,
				Store:        tt.fields.Store,
				StoreTimeout: tt.fields.StoreTimeout,
			}
			if got := j.CanMutate(tt.args.ctx, tt.args.ref); got != tt.want {
				t.Errorf("Javascript.CanMutate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJavascript_Mutate(t *testing.T) {
	type fields struct {
		AssetGetter  asset.Getter
		Store        store.Store
		StoreTimeout time.Duration
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
			j := &Javascript{
				AssetGetter:  tt.fields.AssetGetter,
				Store:        tt.fields.Store,
				StoreTimeout: tt.fields.StoreTimeout,
			}
			got, err := j.Mutate(tt.args.ctx, tt.args.ref, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("Javascript.Mutate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Javascript.Mutate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJavascript_run(t *testing.T) {
	type fields struct {
		AssetGetter  asset.Getter
		Store        store.Store
		StoreTimeout time.Duration
	}
	type args struct {
		ctx     context.Context
		mutator *corev2.Mutator
		event   *corev2.Event
		assets  js.JavascriptAssets
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantFn     func(*corev2.Event) []byte
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "successfully mutates an event when return is not explicitly called in eval",
			args: args{
				ctx: context.Background(),
				mutator: &corev2.Mutator{
					ObjectMeta: corev2.ObjectMeta{
						Namespace: "default",
						Name:      "my_mutator",
					},
					Eval:    `event.check.labels["hockey"] = hockey`,
					Type:    corev2.JavascriptMutator,
					EnvVars: []string{"hockey=puck"},
				},
				event: corev2.FixtureEvent("default", "default"),
			},
			wantFn: func(event *corev2.Event) []byte {
				event.Check.Labels["hockey"] = "puck"
				bytes, _ := json.Marshal(event)
				return bytes
			},
			wantErr: false,
		},
		{
			name: "successfully mutates an event when json stringified event is explicitly returned",
			args: args{
				ctx: context.Background(),
				mutator: &corev2.Mutator{
					ObjectMeta: corev2.ObjectMeta{
						Namespace: "default",
						Name:      "my_mutator",
					},
					Eval:    `event.check.labels["hockey"] = hockey; return JSON.stringify(event);`,
					Type:    corev2.JavascriptMutator,
					EnvVars: []string{"hockey=puck"},
				},
				event: corev2.FixtureEvent("default", "default"),
			},
			wantFn: func(event *corev2.Event) []byte {
				event.Check.Labels["hockey"] = "puck"
				bytes, _ := json.Marshal(event)
				return bytes
			},
			wantErr: false,
		},
		{
			name: "successfully mutates an event when null is explicitly returned",
			args: args{
				ctx: context.Background(),
				mutator: &corev2.Mutator{
					ObjectMeta: corev2.ObjectMeta{
						Namespace: "default",
						Name:      "my_mutator",
					},
					Eval:    `event.check.labels["hockey"] = hockey; return null;`,
					Type:    corev2.JavascriptMutator,
					EnvVars: []string{"hockey=puck"},
				},
				event: corev2.FixtureEvent("default", "default"),
			},
			wantFn: func(event *corev2.Event) []byte {
				event.Check.Labels["hockey"] = "puck"
				bytes, _ := json.Marshal(event)
				return bytes
			},
			wantErr: false,
		},
		{
			name: "returns an error if eval returns an object",
			args: args{
				ctx: context.Background(),
				mutator: &corev2.Mutator{
					ObjectMeta: corev2.ObjectMeta{
						Namespace: "default",
						Name:      "my_mutator",
					},
					Eval:    `event.check.labels["hockey"] = hockey; return {};`,
					Type:    corev2.JavascriptMutator,
					EnvVars: []string{"hockey=puck"},
				},
				event: corev2.FixtureEvent("default", "default"),
			},
			wantErr:    true,
			wantErrMsg: `bad mutator result: got "Object", want string or undefined`,
		},
		{
			name: "returns an error if the timeout is reached",
			args: args{
				ctx: context.Background(),
				mutator: &corev2.Mutator{
					ObjectMeta: corev2.ObjectMeta{
						Namespace: "default",
						Name:      "my_mutator",
					},
					Eval:    `while(true){}`,
					Type:    corev2.JavascriptMutator,
					EnvVars: []string{"hockey=puck"},
					Timeout: 1,
				},
				event: corev2.FixtureEvent("default", "default"),
			},
			wantErr:    true,
			wantErrMsg: "mutator timeout reached, execution halted",
		},
		{
			name: "successfuly mutates when assets are specified",
			args: args{
				ctx: context.Background(),
				mutator: &corev2.Mutator{
					ObjectMeta: corev2.ObjectMeta{
						Namespace: "default",
						Name:      "my_mutator",
					},
					Eval:          `assetFunc();`,
					Type:          corev2.JavascriptMutator,
					EnvVars:       []string{"hockey=puck"},
					RuntimeAssets: []string{"mutatorAsset"},
				},
				event:  corev2.FixtureEvent("default", "default"),
				assets: mutatorAssetSet{},
			},
			wantFn: func(event *corev2.Event) []byte {
				event.Check.Labels["hockey"] = "puck"
				bytes, _ := json.Marshal(event)
				return bytes
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &Javascript{
				AssetGetter:  tt.fields.AssetGetter,
				Store:        tt.fields.Store,
				StoreTimeout: tt.fields.StoreTimeout,
			}
			got, err := j.run(tt.args.ctx, tt.args.mutator, tt.args.event, tt.args.assets)
			if (err != nil) != tt.wantErr {
				t.Errorf("Javascript.run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("Javascript.run() errorMsg = %v, wantErrMsg %v", err, tt.wantErrMsg)
				return
			}
			if !tt.wantErr {
				want := tt.wantFn(tt.args.event)
				assert.JSONEq(t, string(got), string(want))
			}
		})
	}
}
