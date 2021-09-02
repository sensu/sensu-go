package pipeline

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/mock"
)

type mockExec struct {
	mock.Mock
}

// func (m *mockExec) HandleEvent(evt *corev2.Event, mut []byte) (rpc.HandleEventResponse, error) {
// 	args := m.Called(evt, mut)
// 	return args.Get(0).(rpc.HandleEventResponse), args.Error(1)
// }

func (m *mockExec) MutateEvent(evt *corev2.Event) ([]byte, error) {
	args := m.Called(evt)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockExec) FilterEvent(evt *corev2.Event) (bool, error) {
	args := m.Called(evt)
	return args.Get(0).(bool), args.Error(1)
}

// No need to override this method, consumers only log its error
func (m *mockExec) Close() error {
	return nil
}

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
