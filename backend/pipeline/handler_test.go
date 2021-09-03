package pipeline

import (
	"context"
	"reflect"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
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

// func TestPipelineExpandHandlers(t *testing.T) {
// 	type storeFunc func(*mockstore.MockStore)

// 	var nilHandler *corev2.Handler
// 	pipeHandler := corev2.FixtureHandler("pipeHandler")
// 	setHandler := &corev2.Handler{
// 		ObjectMeta: corev2.NewObjectMeta("setHandler", "default"),
// 		Type:       corev2.HandlerSetType,
// 		Handlers:   []string{"pipeHandler"},
// 	}
// 	nestedHandler := &corev2.Handler{
// 		ObjectMeta: corev2.NewObjectMeta("nestedHandler", "default"),
// 		Type:       corev2.HandlerSetType,
// 		Handlers:   []string{"setHandler"},
// 	}
// 	recursiveLoopHandler := &corev2.Handler{
// 		ObjectMeta: corev2.NewObjectMeta("recursiveLoopHandler", "default"),
// 		Type:       corev2.HandlerSetType,
// 		Handlers:   []string{"recursiveLoopHandler"},
// 	}

// 	tests := []struct {
// 		name      string
// 		handlers  []string
// 		storeFunc storeFunc
// 		want      map[string]*corev2.Handler
// 	}{
// 		{
// 			name:     "pipe handler",
// 			handlers: []string{"pipeHandler"},
// 			storeFunc: func(s *mockstore.MockStore) {
// 				s.On("GetHandlerByName", mock.Anything, "pipeHandler").Return(pipeHandler, nil)
// 			},
// 			want: map[string]*corev2.Handler{
// 				"pipeHandler": pipeHandler,
// 			},
// 		},
// 		{
// 			name:     "store error",
// 			handlers: []string{"pipeHandler"},
// 			storeFunc: func(s *mockstore.MockStore) {
// 				s.On("GetHandlerByName", mock.Anything, "pipeHandler").Return(nilHandler, errors.New("error"))
// 			},
// 			want: map[string]*corev2.Handler{},
// 		},
// 		{
// 			name:     "set handler",
// 			handlers: []string{"setHandler"},
// 			storeFunc: func(s *mockstore.MockStore) {
// 				s.On("GetHandlerByName", mock.Anything, "setHandler").Return(setHandler, nil)
// 				s.On("GetHandlerByName", mock.Anything, "pipeHandler").Return(pipeHandler, nil)
// 			},
// 			want: map[string]*corev2.Handler{
// 				"pipeHandler": pipeHandler,
// 			},
// 		},
// 		{
// 			name:     "too deeply nested set handler",
// 			handlers: []string{"recursiveLoopHandler"},
// 			storeFunc: func(s *mockstore.MockStore) {
// 				s.On("GetHandlerByName", mock.Anything, "recursiveLoopHandler").Return(recursiveLoopHandler, nil)
// 			},
// 			want: map[string]*corev2.Handler{},
// 		},
// 		{
// 			name:     "multiple nested set handlers",
// 			handlers: []string{"recursiveLoopHandler", "nestedHandler"},
// 			storeFunc: func(s *mockstore.MockStore) {
// 				s.On("GetHandlerByName", mock.Anything, "recursiveLoopHandler").Return(recursiveLoopHandler, nil)
// 				s.On("GetHandlerByName", mock.Anything, "nestedHandler").Return(nestedHandler, nil)
// 				s.On("GetHandlerByName", mock.Anything, "setHandler").Return(setHandler, nil)
// 				s.On("GetHandlerByName", mock.Anything, "pipeHandler").Return(pipeHandler, nil)
// 			},
// 			want: map[string]*corev2.Handler{
// 				"pipeHandler": pipeHandler,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			store := &mockstore.MockStore{}
// 			if tt.storeFunc != nil {
// 				tt.storeFunc(store)
// 			}

// 			p := &Pipeline{store: store}
// 			got, _ := p.expandHandlers(context.Background(), tt.handlers, 1)
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Pipeline.expandHandlers() = %#v, want %#v", got, tt.want)
// 			}
// 		})
// 	}
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

func TestAdapterV1_getHandlerForResource(t *testing.T) {
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
		name    string
		fields  fields
		args    args
		want    HandlerAdapter
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
			got, err := a.getHandlerForResource(tt.args.ctx, tt.args.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("AdapterV1.getHandlerForResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AdapterV1.getHandlerForResource() = %v, want %v", got, tt.want)
			}
		})
	}
}
