// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"encoding/json"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelined(t *testing.T) {
	p := &Pipelined{}

	bus := &messaging.WizardBus{}
	require.NoError(t, bus.Start())
	p.MessageBus = bus

	store := &mockstore.MockStore{}
	p.Store = store

	assert.NoError(t, p.Start())

	entity := types.FixtureEntity("entity1")
	check := types.FixtureCheck("check1")
	metrics := types.FixtureMetrics()

	event := &types.Event{
		Entity:  entity,
		Check:   check,
		Metrics: metrics,
	}

	notIncident, _ := json.Marshal(event)
	assert.NoError(t, bus.Publish(messaging.TopicEvent, notIncident))

	event.Check.Status = 1

	incident, _ := json.Marshal(event)
	assert.NoError(t, bus.Publish(messaging.TopicEvent, incident))

	assert.NoError(t, p.Stop())
}
