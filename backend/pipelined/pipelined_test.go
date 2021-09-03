// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelined(t *testing.T) {
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start())

	p, err := New(Config{Bus: bus})
	require.NoError(t, err)
	require.NoError(t, p.Start())

	entity := corev2.FixtureEntity("entity1")
	check := corev2.FixtureCheck("check1")
	metrics := corev2.FixtureMetrics()

	event := &corev2.Event{
		Entity:  entity,
		Check:   check,
		Metrics: metrics,
	}

	assert.NoError(t, bus.Publish(messaging.TopicEvent, event))

	event.Check.Status = 1
	assert.NoError(t, bus.Publish(messaging.TopicEvent, event))

	assert.NoError(t, p.Stop())
}
