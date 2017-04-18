// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"testing"

	"github.com/sensu/sensu-go/testing/fixtures"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestPipelinedHandleEvent(t *testing.T) {
	p := &Pipelined{}

	store := fixtures.NewFixtureStore()
	p.Store = store

	entity, _ := store.GetEntityByID("entity1")
	check, _ := store.GetCheckByName("check1")

	event := &types.Event{
		Entity: entity,
		Check:  check,
	}

	// Currently fire and forget. You may choose to return a map
	// of handler execution information in the future, don't know
	// how useful this would be.
	assert.NoError(t, p.handleEvent(event))

	event.Check.Handlers = []string{"handler6"}

	assert.NoError(t, p.handleEvent(event))
}

func TestPipelinedExpandHandlers(t *testing.T) {
	p := &Pipelined{}

	store := fixtures.NewFixtureStore()
	p.Store = store

	oneLevel, err := p.expandHandlers([]string{"handler1"}, 1)

	assert.NoError(t, err)

	handler1, _ := store.GetHandlerByName("handler1")
	expanded := map[string]*types.Handler{"handler1": handler1}

	assert.Equal(t, expanded, oneLevel)

	twoLevels, err := p.expandHandlers([]string{"handler3"}, 1)

	assert.NoError(t, err)

	assert.Equal(t, expanded, twoLevels)

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
