// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"encoding/json"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestPipelinedHandle(t *testing.T) {
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
