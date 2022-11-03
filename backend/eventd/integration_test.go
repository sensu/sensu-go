//go:build integration
// +build integration

package eventd

import (
	"context"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/seeds"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	otherTestutil "github.com/sensu/sensu-go/testing/testutil"
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

	client := ed.NewEmbeddedClient()

	livenessFactory := liveness.EtcdFactory(context.Background(), client)

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
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

	store, err := testutil.NewStoreInstance()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	storev2 := etcdstore.NewStore(store.Client)

	if err := seeds.SeedInitialData(store); err != nil {
		assert.FailNow(t, err.Error())
	}

	e := newEventd(storev2, store, bus, livenessFactory)

	if err := e.Start(); err != nil {
		assert.FailNow(t, err.Error())
	}

	event := corev2.FixtureEvent("entity1", "check1")
	event.Check.Interval = 1
	event.Check.Ttl = 5

	ctx := otherTestutil.ContextWithNamespace("default")(context.Background())

	if err := store.UpdateEntity(ctx, event.Entity); err != nil {
		t.Fatal(err)
	}

	if err := bus.Publish(messaging.TopicEventRaw, event); err != nil {
		assert.FailNow(t, "failed to publish event to TopicEventRaw")
	}

	msg, ok := <-eventChan
	if !ok {
		assert.FailNow(t, "failed to pull message off eventChan")
	}

	okEvent, ok := msg.(*corev2.Event)
	if !ok {
		assert.FailNow(t, "message type was not an event")
	}
	assert.Equal(t, uint32(0), okEvent.Check.Status)

	msg, ok = <-eventChan
	if !ok {
		assert.FailNow(t, "failed to pull message off eventChan")
	}
	warnEvent, ok := msg.(*corev2.Event)
	if !ok {
		assert.FailNow(t, "message type was not an event")
	}
	assert.Equal(t, uint32(1), warnEvent.Check.Status)

	assert.NoError(t, sub.Cancel())
	close(eventChan)
}
