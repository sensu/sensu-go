package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"strings"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store/cache"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/rpc"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockExec struct {
	mock.Mock
}

func (m *mockExec) HandleEvent(evt *corev2.Event, mut []byte) (rpc.HandleEventResponse, error) {
	args := m.Called(evt, mut)
	return args.Get(0).(rpc.HandleEventResponse), args.Error(1)
}

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

func TestHelperHandlerProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_HANDLER_PROCESS") != "1" {
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

func TestPipelineHandleEvent(t *testing.T) {
	t.SkipNow()
	p := &Pipeline{}

	store := &mockstore.MockStore{}
	p.store = store

	entity := corev2.FixtureEntity("entity1")
	check := corev2.FixtureCheck("check1")
	handler := corev2.FixtureHandler("handler1")
	handler.Type = "udp"
	handler.Socket = &corev2.HandlerSocket{
		Host: "127.0.0.1",
		Port: 6789,
	}
	event := &corev2.Event{
		Entity: entity,
		Check:  check,
	}
	extension := &corev2.Extension{
		URL: "http://127.0.0.1",
	}

	// Currently fire and forget. You may choose to return a map
	// of handler execution information in the future, don't know
	// how useful this would be.
	assert.NoError(t, p.HandleEvent(context.Background(), event))

	event.Check.Handlers = []string{"handler1", "handler2"}

	store.On("GetHandlerByName", mock.Anything, "handler1").Return(handler, nil)
	store.On("GetHandlerByName", mock.Anything, "handler2").Return((*corev2.Handler)(nil), nil)
	store.On("GetExtension", mock.Anything, "handler2").Return(extension, nil)
	m := &mockExec{}
	m.On("HandleEvent", event, mock.Anything).Return(rpc.HandleEventResponse{
		Output: "ok",
		Error:  "",
	}, nil)
	p.extensionExecutor = func(*corev2.Extension) (rpc.ExtensionExecutor, error) {
		return m, nil
	}

	assert.NoError(t, p.HandleEvent(context.Background(), event))
	m.AssertCalled(t, "HandleEvent", event, mock.Anything)
}

type mockCache struct {
	mock.Mock
}

func (m *mockCache) Get(namespace string) []cache.Value {
	args := m.Called(namespace)
	return args.Get(0).([]cache.Value)
}

func TestPipelineExpandHandlers(t *testing.T) {
	type cacheFunc func(*mockCache)
	type storeFunc func(*mockstore.MockStore)

	pipeHandler := corev2.FixtureHandler("pipeHandler")
	setHandler := &corev2.Handler{
		ObjectMeta: corev2.NewObjectMeta("setHandler", "default"),
		Type:       corev2.HandlerSetType,
		Handlers:   []string{"pipeHandler"},
	}
	nestedHandler := &corev2.Handler{
		ObjectMeta: corev2.NewObjectMeta("nestedHandler", "default"),
		Type:       corev2.HandlerSetType,
		Handlers:   []string{"setHandler"},
	}
	recursiveLoopHandler := &corev2.Handler{
		ObjectMeta: corev2.NewObjectMeta("recursiveLoopHandler", "default"),
		Type:       corev2.HandlerSetType,
		Handlers:   []string{"recursiveLoopHandler"},
	}

	tests := []struct {
		name      string
		handlers  []string
		storeFunc storeFunc
		cacheFunc cacheFunc
		want      map[string]handlerExtensionUnion
	}{
		{
			name:     "pipe handler",
			handlers: []string{"pipeHandler"},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("GetHandlerByName", mock.Anything, "pipeHandler").Return(pipeHandler, nil)
			},
			cacheFunc: func(c *mockCache) {
				c.On("Get", "default").Return([]cache.Value{
					{Resource: pipeHandler},
				})
			},
			want: map[string]handlerExtensionUnion{
				"pipeHandler": {Handler: pipeHandler},
			},
		},
		{
			name:     "set handler",
			handlers: []string{"setHandler"},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("GetHandlerByName", mock.Anything, "setHandler").Return(setHandler, nil)
				s.On("GetHandlerByName", mock.Anything, "pipeHandler").Return(pipeHandler, nil)
			},
			cacheFunc: func(c *mockCache) {
				c.On("Get", "default").Return([]cache.Value{
					{Resource: pipeHandler},
					{Resource: setHandler},
				})
			},
			want: map[string]handlerExtensionUnion{
				"pipeHandler": {Handler: pipeHandler},
			},
		},
		{
			name:     "too deeply nested set handler",
			handlers: []string{"recursiveLoopHandler"},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("GetHandlerByName", mock.Anything, "recursiveLoopHandler").Return(recursiveLoopHandler, nil)
			},
			cacheFunc: func(c *mockCache) {
				c.On("Get", "default").Return([]cache.Value{
					{Resource: recursiveLoopHandler},
				})
			},
			want: map[string]handlerExtensionUnion{},
		},
		{
			name:     "multiple nested set handlers",
			handlers: []string{"recursiveLoopHandler", "nestedHandler"},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("GetHandlerByName", mock.Anything, "recursiveLoopHandler").Return(recursiveLoopHandler, nil)
				s.On("GetHandlerByName", mock.Anything, "nestedHandler").Return(nestedHandler, nil)
				s.On("GetHandlerByName", mock.Anything, "setHandler").Return(setHandler, nil)
				s.On("GetHandlerByName", mock.Anything, "pipeHandler").Return(pipeHandler, nil)
			},
			cacheFunc: func(c *mockCache) {
				c.On("Get", "default").Return([]cache.Value{
					{Resource: recursiveLoopHandler},
					{Resource: nestedHandler},
					{Resource: setHandler},
					{Resource: pipeHandler},
				})
			},
			want: map[string]handlerExtensionUnion{
				"pipeHandler": {Handler: pipeHandler},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockstore.MockStore{}
			if tt.storeFunc != nil {
				tt.storeFunc(store)
			}
			cache := &mockCache{}
			if tt.cacheFunc != nil {
				tt.cacheFunc(cache)
			}

			ctx := context.WithValue(context.Background(), corev2.NamespaceKey, "default")

			p := &Pipeline{store: store, handlersCache: cache}
			got, _ := p.expandHandlers(ctx, tt.handlers, 1)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pipeline.expandHandlers() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestPipelinePipeHandler(t *testing.T) {
	p := &Pipeline{secretsProviderManager: secrets.NewProviderManager()}
	p.executor = &command.ExecutionRequest{}

	handler := corev2.FakeHandlerCommand("cat")
	handler.Type = "pipe"

	event := corev2.FixtureEvent("test", "test")
	eventData, _ := json.Marshal(event)

	handlerExec, err := p.pipeHandler(handler, event, eventData)

	assert.NoError(t, err)
	assert.Equal(t, string(eventData[:]), handlerExec.Output)
	assert.Equal(t, 0, handlerExec.Status)
}

func TestPipelineTcpHandler(t *testing.T) {
	ready := make(chan struct{})
	done := make(chan struct{})

	p := &Pipeline{secretsProviderManager: secrets.NewProviderManager()}

	handlerSocket := &corev2.HandlerSocket{
		Host: "127.0.0.1",
		Port: 5678,
	}

	handler := &corev2.Handler{
		Type:   "tcp",
		Socket: handlerSocket,
	}

	event := corev2.FixtureEvent("test", "test")
	eventData, _ := json.Marshal(event)

	go func() {
		listener, err := net.Listen("tcp", "127.0.0.1:5678")
		assert.NoError(t, err)
		if err != nil {
			return
		}

		defer func() {
			require.NoError(t, listener.Close())
		}()

		ready <- struct{}{}

		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() {
			require.NoError(t, conn.Close())
		}()

		buffer, err := ioutil.ReadAll(conn)
		if err != nil {
			return
		}

		assert.Equal(t, eventData, buffer)
		done <- struct{}{}
	}()

	<-ready
	_, err := p.socketHandler(handler, event, eventData)

	assert.NoError(t, err)
	<-done
}

func TestPipelineUdpHandler(t *testing.T) {
	ready := make(chan struct{})
	done := make(chan struct{})

	p := &Pipeline{}

	handlerSocket := &corev2.HandlerSocket{
		Host: "127.0.0.1",
		Port: 5678,
	}

	handler := &corev2.Handler{
		Type:   "udp",
		Socket: handlerSocket,
	}

	event := corev2.FixtureEvent("test", "test")
	eventData, _ := json.Marshal(event)

	go func() {
		listener, err := net.ListenPacket("udp", ":5678")
		assert.NoError(t, err)
		if err != nil {
			return
		}

		defer func() {
			require.NoError(t, listener.Close())
		}()

		ready <- struct{}{}

		buffer := make([]byte, 8192)
		rlen, _, err := listener.ReadFrom(buffer)

		assert.NoError(t, err)
		assert.Equal(t, eventData, buffer[0:rlen])
		done <- struct{}{}
	}()

	<-ready

	_, err := p.socketHandler(handler, event, eventData)

	assert.NoError(t, err)
	<-done
}

func TestPipelineGRPCHandler(t *testing.T) {
	extension := &corev2.Extension{}
	event := corev2.FixtureEvent("foo", "bar")
	execFn := func(ext *corev2.Extension) (rpc.ExtensionExecutor, error) {
		mock := &mockExec{}
		mock.On("HandleEvent", event, []byte(nil)).Return(rpc.HandleEventResponse{
			Output: "ok",
			Error:  "",
		}, nil)
		return mock, nil
	}
	p := &Pipeline{
		extensionExecutor: execFn,
	}
	result, err := p.grpcHandler(extension, event, nil)

	assert.NoError(t, err)
	assert.Equal(t, "ok", result.Output)
	assert.Equal(t, "", result.Error)
}
