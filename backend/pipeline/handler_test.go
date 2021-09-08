package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockpipeline"
	"github.com/stretchr/testify/mock"
)

func TestAdapterV1_processHandler(t *testing.T) {
	type fields struct {
		Store           store.Store
		StoreTimeout    time.Duration
		FilterAdapters  []FilterAdapter
		MutatorAdapters []MutatorAdapter
		HandlerAdapters []HandlerAdapter
	}
	type args struct {
		ctx         context.Context
		ref         *corev2.ResourceReference
		event       *corev2.Event
		mutatedData []byte
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns an error when getHandlerAdapterForResource() returns an error",
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Handler",
					Name:       "handler1",
				},
			},
			wantErr:    true,
			wantErrMsg: "no handler adapters were found that can handle the resource: core/v2.Handler = handler1",
		},
		{
			name: "returns an error when handler.Handle() returns an error",
			fields: fields{
				HandlerAdapters: func() []HandlerAdapter {
					adapter := mockpipeline.HandlerAdapter{}
					adapter.On("CanHandle", mock.Anything).Return(true)
					adapter.On("Handle", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(errors.New("handler error"))
					return []HandlerAdapter{adapter}
				}(),
			},
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Handler",
					Name:       "handler1",
				},
			},
			wantErr:    true,
			wantErrMsg: "handler error",
		},
		{
			name: "returns nil when no errors occur",
			fields: fields{
				HandlerAdapters: func() []HandlerAdapter {
					adapter := mockpipeline.HandlerAdapter{}
					adapter.On("CanHandle", mock.Anything).Return(true)
					adapter.On("Handle", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(nil)
					return []HandlerAdapter{adapter}
				}(),
			},
			args: args{
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Handler",
					Name:       "handler1",
				},
			},
			wantErr: false,
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
			err := a.processHandler(tt.args.ctx, tt.args.ref, tt.args.event, tt.args.mutatedData)
			if (err != nil) != tt.wantErr {
				t.Errorf("AdapterV1.processHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("AdapterV1.processHandler() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
			}
		})
	}
}

func TestAdapterV1_getHandlerAdapterForResource(t *testing.T) {
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
			name: "returns an error when no handler adapters exist",
			args: args{
				ctx: context.Background(),
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Handler",
					Name:       "handler1",
				},
			},
			wantNil:    true,
			wantErr:    true,
			wantErrMsg: "no handler adapters were found that can handle the resource: core/v2.Handler = handler1",
		},
		{
			name: "returns an error when no handler adapters support the resource reference",
			fields: fields{
				HandlerAdapters: func() []HandlerAdapter {
					adapter := mockpipeline.HandlerAdapter{}
					adapter.On("CanHandle", mock.Anything).Return(false)
					return []HandlerAdapter{adapter}
				}(),
			},
			args: args{
				ctx: context.Background(),
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Handler",
					Name:       "handler1",
				},
			},
			wantNil:    true,
			wantErr:    true,
			wantErrMsg: "no handler adapters were found that can handle the resource: core/v2.Handler = handler1",
		},
		{
			name: "returns the first adapter that can support the resource reference",
			fields: fields{
				HandlerAdapters: func() []HandlerAdapter {
					adapter1 := mockpipeline.HandlerAdapter{}
					adapter1.On("Name").Return("adapter1")
					adapter1.On("CanHandle", mock.Anything).Return(false)

					adapter2 := mockpipeline.HandlerAdapter{}
					adapter2.On("Name").Return("adapter2")
					adapter2.On("CanHandle", mock.Anything).Return(true)

					adapter3 := mockpipeline.HandlerAdapter{}
					adapter3.On("Name").Return("adapter3")
					adapter3.On("CanHandle", mock.Anything).Return(true)

					return []HandlerAdapter{adapter1, adapter2, adapter3}
				}(),
			},
			args: args{
				ctx: context.Background(),
				ref: &corev2.ResourceReference{
					APIVersion: "core/v2",
					Type:       "Handler",
					Name:       "handler1",
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
			got, err := a.getHandlerAdapterForResource(tt.args.ctx, tt.args.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("AdapterV1.getHandlerAdapterForResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("AdapterV1.getHandlerAdapterForResource() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}
			gotNil := (got == nil)
			if gotNil != tt.wantNil {
				t.Errorf("AdapterV1.getHandlerAdapterForResource() nil = %v, wantNil %v", gotNil, tt.wantNil)
				return
			}
			if got != nil {
				if got.Name() != tt.wantName {
					t.Errorf("AdapterV1.getHandlerAdapterForResource() Name() = %v, wantName %v", got, tt.wantName)
				}
			}
		})
	}
}
