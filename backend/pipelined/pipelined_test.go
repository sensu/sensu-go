// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"encoding/json"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestPipelined(t *testing.T) {
	p := &Pipelined{}

	bus := &messaging.WizardBus{}
	bus.Start()

	p.MessageBus = bus

	assert.NoError(t, p.Start())

	check := &types.Check{}
	check.Handlers = []string{"debug"}

	event := &types.Event{}
	event.Check = check

	eventJSON, _ := json.Marshal(event)

	err := bus.Publish("sensu:event", eventJSON)
	assert.NoError(t, err)

	assert.NoError(t, p.Stop())
}
