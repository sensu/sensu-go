// +build integration,race

package keepalived

import (
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/seeds"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeepaliveMonitor(t *testing.T) {
	bus := &messaging.WizardBus{}

	if err := bus.Start(); err != nil {
		assert.FailNow(t, "message bus failed to start")
	}

	eventChan := make(chan interface{}, 2)

	if err := bus.Subscribe(messaging.TopicEventRaw, "test", eventChan); err != nil {
		assert.FailNow(t, "failed to subscribe to message bus topic event raw")
	}

	store, err := testutil.NewStoreInstance()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	if err := seeds.SeedInitialData(store); err != nil {
		assert.FailNow(t, err.Error())
	}

	k, err := New(Config{Store: store, Bus: bus})
	require.NoError(t, err)

	if err := k.Start(); err != nil {
		assert.FailNow(t, err.Error())
	}

	entity := types.FixtureEntity("entity1")
	entity.KeepaliveTimeout = 1

	keepalive := &types.Event{
		Entity:    entity,
		Timestamp: time.Now().Unix(),
	}

	if err := bus.Publish(messaging.TopicKeepalive, keepalive); err != nil {
		assert.FailNow(t, "failed to publish keepalive event")
	}

	msg, ok := <-eventChan
	if !ok {
		assert.FailNow(t, "failed to pull message off eventChan")
	}

	okEvent, ok := msg.(*types.Event)
	if !ok {
		assert.FailNow(t, "message type was not an event")
	}
	assert.Equal(t, uint32(0), okEvent.Check.Status)

	msg, ok = <-eventChan
	if !ok {
		assert.FailNow(t, "failed to pull message off eventChan")
	}
	warnEvent, ok := msg.(*types.Event)
	if !ok {
		assert.FailNow(t, "message type was not an event")
	}
	assert.Equal(t, uint32(1), warnEvent.Check.Status)
}
