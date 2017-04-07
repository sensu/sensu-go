// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestPipelinedFilter(t *testing.T) {
	p := &Pipelined{}

	handler := &types.Handler{
		Type:    "pipe",
		Command: "cat",
	}

	event := &types.Event{
		Check: &types.Check{},
	}

	event.Check.Status = 0

	notIncident := p.filterEvent(handler, event)

	assert.Equal(t, true, notIncident)

	event.Check.Status = 1

	incident := p.filterEvent(handler, event)

	assert.Equal(t, false, incident)

	event.Check.Status = 0

	noMetrics := p.filterEvent(handler, event)

	assert.Equal(t, true, noMetrics)

	event.Metrics = &types.Metrics{}

	metrics := p.filterEvent(handler, event)

	assert.Equal(t, false, metrics)

	event.Check.Status = 1

	both := p.filterEvent(handler, event)

	assert.Equal(t, false, both)
}

func TestPipelinedIsIncident(t *testing.T) {
	p := &Pipelined{}

	event := &types.Event{
		Check: &types.Check{},
	}

	notIncident := p.isIncident(event)

	assert.Equal(t, false, notIncident)

	event.Check.Status = 1

	incident := p.isIncident(event)

	assert.Equal(t, true, incident)
}

func TestPipelinedHasMetrics(t *testing.T) {
	p := &Pipelined{}

	event := &types.Event{
		Check: &types.Check{},
	}

	notMetrics := p.hasMetrics(event)

	assert.Equal(t, false, notMetrics)

	event.Metrics = &types.Metrics{}

	metrics := p.hasMetrics(event)

	assert.Equal(t, true, metrics)
}
