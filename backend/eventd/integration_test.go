// +build integration

package eventd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/monitor"
	"github.com/sensu/sensu-go/backend/seeds"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/sensu/sensu-go/testing/mockring"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testReceiver struct {
	c chan interface{}
}

func (r testReceiver) Receiver() chan<- interface{} {
	return r.c
}

func TestEventdMonitor(t *testing.T) {
	ed, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := ed.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	monFac := monitor.EtcdFactory(client)

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{
		RingGetter: &mockring.Getter{},
	})
	require.NoError(t, err)

	if err := bus.Start(); err != nil {
		assert.FailNow(t, "message bus failed to start")
	}

	eventChan := make(chan interface{}, 2)

	subscriber := testReceiver{
		c: eventChan,
	}
	sub, err := bus.Subscribe(messaging.TopicEvent, "testReceiver", subscriber)
	if err != nil {
		assert.FailNow(t, "failed to subscribe to message bus topic event")
	}

	storeInst, err := testutil.NewStoreInstance()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	store := storeInst.GetStore()
	storev2 := storeInst.GetStoreV2()
	if err := seeds.SeedInitialData(store, storev2); err != nil {
		assert.FailNow(t, err.Error())
	}

	e, err := New(Config{Store: store, Bus: bus, MonitorFactory: monFac})
	require.NoError(t, err)

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

	assert.NoError(t, sub.Cancel())
	close(eventChan)
}
