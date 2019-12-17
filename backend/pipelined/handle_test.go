// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"strings"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/rpc"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockExec struct {
	mock.Mock
}

func (m *mockExec) HandleEvent(evt *types.Event, mut []byte) (rpc.HandleEventResponse, error) {
	args := m.Called(evt, mut)
	return args.Get(0).(rpc.HandleEventResponse), args.Error(1)
}

func (m *mockExec) MutateEvent(evt *types.Event) ([]byte, error) {
	args := m.Called(evt)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockExec) FilterEvent(evt *types.Event) (bool, error) {
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

func TestPipelinedHandleEvent(t *testing.T) {
	t.SkipNow()
	p := &Pipelined{}

	store := &mockstore.MockStore{}
	p.store = store

	entity := types.FixtureEntity("entity1")
	check := types.FixtureCheck("check1")
	handler := types.FixtureHandler("handler1")
	handler.Type = "udp"
	handler.Socket = &types.HandlerSocket{
		Host: "127.0.0.1",
		Port: 6789,
	}
	event := &types.Event{
		Entity: entity,
		Check:  check,
	}
	extension := &types.Extension{
		URL: "http://127.0.0.1",
	}

	// Currently fire and forget. You may choose to return a map
	// of handler execution information in the future, don't know
	// how useful this would be.
	assert.NoError(t, p.handleEvent(event))

	event.Check.Handlers = []string{"handler1", "handler2"}

	store.On("GetHandlerByName", mock.Anything, "handler1").Return(handler, nil)
	store.On("GetHandlerByName", mock.Anything, "handler2").Return((*types.Handler)(nil), nil)
	store.On("GetExtension", mock.Anything, "handler2").Return(extension, nil)
	m := &mockExec{}
	m.On("HandleEvent", event, mock.Anything).Return(rpc.HandleEventResponse{
		Output: "ok",
		Error:  "",
	}, nil)
	p.extensionExecutor = func(*types.Extension) (rpc.ExtensionExecutor, error) {
		return m, nil
	}

	assert.NoError(t, p.handleEvent(event))
	m.AssertCalled(t, "HandleEvent", event, mock.Anything)
}

func TestPipelinedExpandHandlers(t *testing.T) {
	type storeFunc func(*mockstore.MockStore)

	var nilHandler *corev2.Handler
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
		want      map[string]handlerExtensionUnion
	}{
		{
			name:     "pipe handler",
			handlers: []string{"pipeHandler"},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("GetHandlerByName", mock.Anything, "pipeHandler").Return(pipeHandler, nil)
			},
			want: map[string]handlerExtensionUnion{
				"pipeHandler": handlerExtensionUnion{Handler: pipeHandler},
			},
		},
		{
			name:     "store error",
			handlers: []string{"pipeHandler"},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("GetHandlerByName", mock.Anything, "pipeHandler").Return(nilHandler, errors.New("error"))
			},
			want: map[string]handlerExtensionUnion{},
		},
		{
			name:     "set handler",
			handlers: []string{"setHandler"},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("GetHandlerByName", mock.Anything, "setHandler").Return(setHandler, nil)
				s.On("GetHandlerByName", mock.Anything, "pipeHandler").Return(pipeHandler, nil)
			},
			want: map[string]handlerExtensionUnion{
				"pipeHandler": handlerExtensionUnion{Handler: pipeHandler},
			},
		},
		{
			name:     "too deeply nested set handler",
			handlers: []string{"recursiveLoopHandler"},
			storeFunc: func(s *mockstore.MockStore) {
				s.On("GetHandlerByName", mock.Anything, "recursiveLoopHandler").Return(recursiveLoopHandler, nil)
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
			want: map[string]handlerExtensionUnion{
				"pipeHandler": handlerExtensionUnion{Handler: pipeHandler},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockstore.MockStore{}
			if tt.storeFunc != nil {
				tt.storeFunc(store)
			}

			p := &Pipelined{store: store}
			got, _ := p.expandHandlers(context.Background(), tt.handlers, 1)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pipelined.expandHandlers() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestPipelinedPipeHandler(t *testing.T) {
	p := &Pipelined{secretsProviderManager: secrets.NewProviderManager()}
	p.executor = &command.ExecutionRequest{}

	handler := types.FakeHandlerCommand("cat")
	handler.Type = "pipe"

	event := &types.Event{}
	eventData, _ := json.Marshal(event)

	handlerExec, err := p.pipeHandler(handler, eventData)

	assert.NoError(t, err)
	assert.Equal(t, string(eventData[:]), handlerExec.Output)
	assert.Equal(t, 0, handlerExec.Status)
}

func TestPipelinedTcpHandler(t *testing.T) {
	ready := make(chan struct{})
	done := make(chan struct{})

	p := &Pipelined{secretsProviderManager: secrets.NewProviderManager()}

	handlerSocket := &types.HandlerSocket{
		Host: "127.0.0.1",
		Port: 5678,
	}

	handler := &types.Handler{
		Type:   "tcp",
		Socket: handlerSocket,
	}

	event := &types.Event{}
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
	_, err := p.socketHandler(handler, eventData)

	assert.NoError(t, err)
	<-done
}

func TestPipelinedUdpHandler(t *testing.T) {
	ready := make(chan struct{})
	done := make(chan struct{})

	p := &Pipelined{}

	handlerSocket := &types.HandlerSocket{
		Host: "127.0.0.1",
		Port: 5678,
	}

	handler := &types.Handler{
		Type:   "udp",
		Socket: handlerSocket,
	}

	event := &types.Event{}
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

		buffer := make([]byte, 1024)
		rlen, _, err := listener.ReadFrom(buffer)

		assert.NoError(t, err)
		assert.Equal(t, eventData, buffer[0:rlen])
		done <- struct{}{}
	}()

	<-ready

	_, err := p.socketHandler(handler, eventData)

	assert.NoError(t, err)
	<-done
}

func TestPipelinedGRPCHandler(t *testing.T) {
	extension := &types.Extension{}
	event := types.FixtureEvent("foo", "bar")
	execFn := func(ext *types.Extension) (rpc.ExtensionExecutor, error) {
		mock := &mockExec{}
		mock.On("HandleEvent", event, []byte(nil)).Return(rpc.HandleEventResponse{
			Output: "ok",
			Error:  "",
		}, nil)
		return mock, nil
	}
	p := &Pipelined{
		extensionExecutor: execFn,
	}
	result, err := p.grpcHandler(extension, event, nil)

	assert.NoError(t, err)
	assert.Equal(t, "ok", result.Output)
	assert.Equal(t, "", result.Error)
}
