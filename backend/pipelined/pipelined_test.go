// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"encoding/json"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/fixtures"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestPipelined(t *testing.T) {
	p := &Pipelined{}

	bus := &messaging.WizardBus{}
	bus.Start()
	p.MessageBus = bus

	store := fixtures.NewFixtureStore()
	p.Store = store

	assert.NoError(t, p.Start())

	entity, _ := store.GetEntityByID("entity1")
	check, _ := store.GetCheckByName("check1")

	event := &types.Event{
		Entity: entity,
		Check:  check,
	}

	notIncident, _ := json.Marshal(event)
	assert.NoError(t, bus.Publish("sensu:event", notIncident))

	event.Check.Status = 1

	incident, _ := json.Marshal(event)
	assert.NoError(t, bus.Publish("sensu:event", incident))

	assert.NoError(t, p.Stop())
}
