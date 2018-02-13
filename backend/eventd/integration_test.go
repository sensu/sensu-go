// +build integration,race

package eventd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/seeds"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestEventdMonitor(t *testing.T) {
	bus := &messaging.WizardBus{}

	if err := bus.Start(); err != nil {
		assert.FailNow(t, "message bus failed to start")
	}

	eventChan := make(chan interface{}, 2)

	if err := bus.Subscribe(messaging.TopicEvent, "test", eventChan); err != nil {
		assert.FailNow(t, "failed to subscribe to message bus topic event")
	}

	store, err := testutil.NewStoreInstance()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	if err := seeds.SeedInitialData(store); err != nil {
		assert.FailNow(t, err.Error())
	}

	e := &Eventd{
		Store:      store,
		MessageBus: bus,
	}

	if err := e.Start(); err != nil {
		assert.FailNow(t, err.Error())
	}

	event := types.FixtureEvent("entity1", "check1")
	event.Check.Interval = 1
	event.Check.Ttl = 2

	if err := bus.Publish(messaging.TopicEventRaw, event); err != nil {
		assert.FailNow(t, "failed to publish event to TopicEventRaw")
	}

	msg, ok := <-eventChan
	if !ok {
		assert.FailNow(t, "failed to pull message off eventChan")
	}

	okEvent, ok := msg.(*types.Event)
	if !ok {
		assert.FailNow(t, "message type was not an event")
	}
	assert.Equal(t, int32(0), okEvent.Check.Status)

	msg, ok = <-eventChan
	if !ok {
		assert.FailNow(t, "failed to pull message off eventChan")
	}
	warnEvent, ok := msg.(*types.Event)
	if !ok {
		assert.FailNow(t, "message type was not an event")
	}
	assert.Equal(t, int32(1), warnEvent.Check.Status)
}
