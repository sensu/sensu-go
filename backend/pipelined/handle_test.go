// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"encoding/json"
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

	event := &types.Event{}
	event.Entity = entity
	event.Check = check

	// Currently fire and forget. You may choose to return a map
	// of handler execution information in the future, don't know
	// how useful this would be.
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

	handlerPipe := &types.HandlerPipe{
		Command: "cat",
	}

	handler := &types.Handler{
		Type: "pipe",
		Pipe: *handlerPipe,
	}

	event := &types.Event{}
	eventData, _ := json.Marshal(event)

	handlerExec, err := p.pipeHandler(handler, eventData)

	assert.NoError(t, err)
	assert.Equal(t, string(eventData[:]), handlerExec.Output)
	assert.Equal(t, 0, handlerExec.Status)
}
