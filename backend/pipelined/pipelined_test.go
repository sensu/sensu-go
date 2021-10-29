// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelined(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start(ctx))

	p, err := New(Config{Bus: bus})
	require.NoError(t, err)
	require.NoError(t, p.Start(ctx))

	event := corev2.FixtureEvent("entity1", "check1")
	event.Metrics = corev2.FixtureMetrics()

	assert.NoError(t, bus.Publish(messaging.TopicEvent, event))

	event.Check.Status = 1
	assert.NoError(t, bus.Publish(messaging.TopicEvent, event))

	assert.NoError(t, p.Stop())
}
