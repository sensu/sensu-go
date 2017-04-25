// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestPipelinedHandleEvent(t *testing.T) {
	p := &Pipelined{}

	store := &mockstore.MockStore{}
	p.Store = store

	entity := types.FixtureEntity("entity1")
	check := types.FixtureCheck("check1")
	handler := types.FixtureHandler("handler1")
	handler.Type = "udp"
	handler.Socket = types.HandlerSocket{
		Host: "localhost",
		Port: 6789,
	}
	event := &types.Event{
		Entity: entity,
		Check:  check,
	}

	// Currently fire and forget. You may choose to return a map
	// of handler execution information in the future, don't know
	// how useful this would be.
	assert.NoError(t, p.handleEvent(event))

	event.Check.Handlers = []string{"handler1"}
	store.On("GetHandlerByName", "handler1").Return(handler, nil)
	assert.NoError(t, p.handleEvent(event))
}

func TestPipelinedExpandHandlers(t *testing.T) {
	p := &Pipelined{}

	store := &mockstore.MockStore{}
	p.Store = store
	handler1 := types.FixtureHandler("handler1")
	store.On("GetHandlerByName", "handler1").Return(handler1, nil)

	oneLevel, err := p.expandHandlers([]string{"handler1"}, 1)
	assert.NoError(t, err)

	expanded := map[string]*types.Handler{"handler1": handler1}
	assert.Equal(t, expanded, oneLevel)

	handler2 := types.FixtureHandler("handler2")
	handler2.Type = "set"
	handler2.Handlers = []string{"handler1", "unknown"}

	handler3 := types.FixtureHandler("handler3")
	handler3.Type = "set"
	handler3.Handlers = []string{"handler1", "handler2"}

	var nilHandler *types.Handler
	store.On("GetHandlerByName", "unknown").Return(nilHandler, nil)
	store.On("GetHandlerByName", "handler2").Return(handler2, nil)
	store.On("GetHandlerByName", "handler3").Return(handler3, nil)

	twoLevels, err := p.expandHandlers([]string{"handler3"}, 1)
	assert.NoError(t, err)
	assert.Equal(t, expanded, twoLevels)

	handler4 := types.FixtureHandler("handler4")
	handler4.Type = "set"
	handler4.Handlers = []string{"handler2", "handler3"}

	store.On("GetHandlerByName", "handler4").Return(handler4, nil)
	threeLevels, err := p.expandHandlers([]string{"handler4"}, 1)

	assert.NoError(t, err)

	assert.Equal(t, expanded, threeLevels)
}

func TestPipelinedPipeHandler(t *testing.T) {
	p := &Pipelined{}

	handler := &types.Handler{
		Type:    "pipe",
		Command: "cat",
	}

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

	p := &Pipelined{}

	handlerSocket := &types.HandlerSocket{
		Host: "localhost",
		Port: 5678,
	}

	handler := &types.Handler{
		Type:   "tcp",
		Socket: *handlerSocket,
	}

	event := &types.Event{}
	eventData, _ := json.Marshal(event)

	go func() {
		listener, err := net.Listen("tcp", ":5678")
		assert.NoError(t, err)
		if err != nil {
			return
		}

		defer listener.Close()

		ready <- struct{}{}

		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

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
		Host: "localhost",
		Port: 5678,
	}

	handler := &types.Handler{
		Type:   "udp",
		Socket: *handlerSocket,
	}

	event := &types.Event{}
	eventData, _ := json.Marshal(event)

	go func() {
		listener, err := net.ListenPacket("udp", ":5678")
		assert.NoError(t, err)
		if err != nil {
			return
		}

		defer listener.Close()

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
