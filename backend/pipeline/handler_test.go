package pipeline

import (
	"context"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockpipeline"
	"github.com/stretchr/testify/mock"
)

// func TestPipelineHandleEvent(t *testing.T) {
// 	t.SkipNow()
// 	p := &Pipeline{}

// 	store := &mockstore.MockStore{}
// 	p.store = store

// 	entity := corev2.FixtureEntity("entity1")
// 	check := corev2.FixtureCheck("check1")
// 	handler := corev2.FixtureHandler("handler1")
// 	handler.Type = "udp"
// 	handler.Socket = &corev2.HandlerSocket{
// 		Host: "127.0.0.1",
// 		Port: 6789,
// 	}
// 	event := &corev2.Event{
// 		Entity: entity,
// 		Check:  check,
// 	}

// 	// Currently fire and forget. You may choose to return a map
// 	// of handler execution information in the future, don't know
// 	// how useful this would be.
// 	assert.NoError(t, p.HandleEvent(context.Background(), event))

// 	event.Check.Handlers = []string{"handler1", "handler2"}

// 	store.On("GetHandlerByName", mock.Anything, "handler1").Return(handler, nil)
// 	store.On("GetHandlerByName", mock.Anything, "handler2").Return((*corev2.Handler)(nil), nil)
// 	m := &mockExec{}
// 	// m.On("HandleEvent", event, mock.Anything).Return(rpc.HandleEventResponse{
// 	// 	Output: "ok",
// 	// 	Error:  "",
// 	// }, nil)

// 	assert.NoError(t, p.HandleEvent(context.Background(), event))
// 	m.AssertCalled(t, "HandleEvent", event, mock.Anything)
// }

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
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
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
			if err := a.processHandler(tt.args.ctx, tt.args.ref, tt.args.event, tt.args.mutatedData); (err != nil) != tt.wantErr {
				t.Errorf("AdapterV1.processHandler() error = %v, wantErr %v", err, tt.wantErr)
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
