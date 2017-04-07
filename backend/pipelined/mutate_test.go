// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"encoding/json"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestPipelinedMutate(t *testing.T) {
	p := &Pipelined{}

	handler := &types.Handler{
		Type:    "pipe",
		Command: "cat",
	}

	event := &types.Event{}

	eventData, err := p.mutateEvent(handler, event)

	expected, _ := json.Marshal(event)

	assert.NoError(t, err)
	assert.Equal(t, expected, eventData)
}

func TestPipelinedJsonMutator(t *testing.T) {
	p := &Pipelined{}

	event := &types.Event{}

	output, err := p.jsonMutator(event)

	expected, _ := json.Marshal(event)

	assert.NoError(t, err)
	assert.Equal(t, expected, output)
}

func TestPipelinedPipeMutator(t *testing.T) {
	p := &Pipelined{}

	mutator := &types.Mutator{
		Command: "cat",
	}

	event := &types.Event{}

	output, err := p.pipeMutator(mutator, event)

	expected, _ := json.Marshal(event)

	assert.NoError(t, err)
	assert.Equal(t, expected, output)
}
