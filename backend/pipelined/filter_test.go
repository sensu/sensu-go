// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestPipelinedFilter(t *testing.T) {
	p := &Pipelined{}

	handlerPipe := &types.HandlerPipe{
		Command: "cat",
	}

	handler := &types.Handler{
		Type: "pipe",
		Pipe: *handlerPipe,
	}

	event := &types.Event{
		Check: &types.Check{},
	}

	filtered := p.filterEvent(handler, event)

	assert.Equal(t, true, filtered)

	event.Check.Status = 1

	notFiltered := p.filterEvent(handler, event)

	assert.Equal(t, false, notFiltered)
}

func TestPipelinedIsIncident(t *testing.T) {
	p := &Pipelined{}

	event := &types.Event{
		Check: &types.Check{},
	}

	notIncident, err := p.isIncident(event)

	assert.NoError(t, err)
	assert.Equal(t, false, notIncident)

	event.Check.Status = 1

	incident, err := p.isIncident(event)

	assert.NoError(t, err)
	assert.Equal(t, true, incident)
}
