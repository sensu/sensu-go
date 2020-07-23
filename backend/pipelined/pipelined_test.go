// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelined(t *testing.T) {
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start())
	store := &mockstore.MockStore{}

	ctx, cancel := context.WithCancel(context.Background())
	p := &Pipelined{
		ctx:         ctx,
		cancel:      cancel,
		store:       store,
		bus:         bus,
		stopping:    make(chan struct{}, 1),
		running:     &atomic.Value{},
		wg:          &sync.WaitGroup{},
		errChan:     make(chan error, 1),
		eventChan:   make(chan interface{}, 1),
		workerCount: 1,
		executor:    command.NewExecutor(),
	}
	require.NoError(t, p.Start())

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
