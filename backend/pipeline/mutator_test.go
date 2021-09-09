package pipeline

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockpipeline"
	"github.com/stretchr/testify/mock"
)

func TestAdapterV1_processMutator(t *testing.T) {
	type fields struct {
		Store           store.Store
		StoreTimeout    time.Duration
		FilterAdapters  []FilterAdapter
		MutatorAdapters []MutatorAdapter
		HandlerAdapters []HandlerAdapter
	}
	type args struct {
		ctx   context.Context
		ref   *corev2.ResourceReference
		event *corev2.Event
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       []byte
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns an error when getMutatorAdapterForResource() returns an error",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Mutator",
					Name:       "mutator1",
				},
			},
			wantErr:    true,
			wantErrMsg: "no mutator adapters were found that can mutate the resource: core/v2.Mutator = mutator1",
		},
		{
			name: "returns an error when mutator.Mutate() returns an error",
			fields: fields{
				MutatorAdapters: func() []MutatorAdapter {
					adapter := &mockpipeline.MutatorAdapter{}
					adapter.On("CanMutate", mock.Anything).Return(true)
					adapter.On("Mutate", mock.Anything, mock.Anything, mock.Anything).
						Return([]byte{}, errors.New("mutator error"))
					return []MutatorAdapter{adapter}
				}(),
			},
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Mutator",
					Name:       "mutator1",
				},
			},
			want:       []byte{},
			wantErr:    true,
			wantErrMsg: "mutator error",
		},
		{
			name: "can return mutated data without errors",
			fields: fields{
				MutatorAdapters: func() []MutatorAdapter {
					adapter := &mockpipeline.MutatorAdapter{}
					adapter.On("CanMutate", mock.Anything).Return(true)
					adapter.On("Mutate", mock.Anything, mock.Anything, mock.Anything).
						Return([]byte("mutated data"), nil)
					return []MutatorAdapter{adapter}
				}(),
			},
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Mutator",
					Name:       "mutator1",
				},
			},
			want: []byte("mutated data"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AdapterV1{
				Store:           tt.fields.Store,
				StoreTimeout:    tt.fields.StoreTimeout,
				FilterAdapters:  tt.fields.FilterAdapters,
				MutatorAdapters: tt.fields.MutatorAdapters,
				HandlerAdapters: tt.fields.HandlerAdapters,
			}
			got, err := a.processMutator(tt.args.ctx, tt.args.ref, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("AdapterV1.processMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("AdapterV1.processMutator() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AdapterV1.processMutator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdapterV1_getMutatorAdapterForResource(t *testing.T) {
	type fields struct {
		Store           store.Store
		StoreTimeout    time.Duration
		FilterAdapters  []FilterAdapter
		MutatorAdapters []MutatorAdapter
		HandlerAdapters []HandlerAdapter
	}
	type args struct {
		ctx context.Context
		ref *corev2.ResourceReference
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantNil    bool
		wantName   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns an error when no mutator adapters exist",
			args: args{
				ctx: context.Background(),
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Mutator",
					Name:       "mutator1",
				},
			},
			wantNil:    true,
			wantErr:    true,
			wantErrMsg: "no mutator adapters were found that can mutate the resource: core/v2.Mutator = mutator1",
		},
		{
			name: "returns an error when no mutator adapters support the resource reference",
			fields: fields{
				MutatorAdapters: func() []MutatorAdapter {
					adapter := &mockpipeline.MutatorAdapter{}
					adapter.On("CanMutate", mock.Anything).Return(false)
					return []MutatorAdapter{adapter}
				}(),
			},
			args: args{
				ctx: context.Background(),
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Mutator",
					Name:       "mutator1",
				},
			},
			wantNil:    true,
			wantErr:    true,
			wantErrMsg: "no mutator adapters were found that can mutate the resource: core/v2.Mutator = mutator1",
		},
		{
			name: "returns the first adapter that can support the resource reference",
			fields: fields{
				MutatorAdapters: func() []MutatorAdapter {
					adapter1 := &mockpipeline.MutatorAdapter{}
					adapter1.On("Name").Return("adapter1")
					adapter1.On("CanMutate", mock.Anything).Return(false)

					adapter2 := &mockpipeline.MutatorAdapter{}
					adapter2.On("Name").Return("adapter2")
					adapter2.On("CanMutate", mock.Anything).Return(true)

					adapter3 := &mockpipeline.MutatorAdapter{}
					adapter3.On("Name").Return("adapter3")
					adapter3.On("CanMutate", mock.Anything).Return(true)

					return []MutatorAdapter{adapter1, adapter2, adapter3}
				}(),
			},
			args: args{
				ctx: context.Background(),
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Mutator",
					Name:       "mutator1",
				},
			},
			wantName: "adapter2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AdapterV1{
				Store:           tt.fields.Store,
				StoreTimeout:    tt.fields.StoreTimeout,
				FilterAdapters:  tt.fields.FilterAdapters,
				MutatorAdapters: tt.fields.MutatorAdapters,
				HandlerAdapters: tt.fields.HandlerAdapters,
			}
			got, err := a.getMutatorAdapterForResource(tt.args.ctx, tt.args.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("AdapterV1.getMutatorAdapterForResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("AdapterV1.getMutatorAdapterForResource() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}
			gotNil := (got == nil)
			if gotNil != tt.wantNil {
				t.Errorf("AdapterV1.getMutatorAdapterForResource() nil = %v, wantNil %v", gotNil, tt.wantNil)
				return
			}
			if got != nil {
				if got.Name() != tt.wantName {
					t.Errorf("AdapterV1.getMutatorAdapterForResource() Name() = %v, wantName %v", got, tt.wantName)
				}
			}
		})
	}
}
