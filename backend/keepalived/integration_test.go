// +build integration

package keepalived

import (
	"context"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/seeds"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeepaliveMonitor(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)

	if err := bus.Start(); err != nil {
		assert.FailNow(t, "message bus failed to start")
	}

	eventChan := make(chan interface{}, 2)

	tsub := testSubscriber{
		ch: eventChan,
	}
	subscription, err := bus.Subscribe(messaging.TopicEventRaw, "testSubscriber", tsub)
	if err != nil {
		assert.FailNow(t, "failed to subscribe to message bus topic event raw")
	}
	defer subscription.Cancel()

	entity := corev2.FixtureEntity("entity1")
	ctx := store.NamespaceContext(context.Background(), entity.Namespace)

	store, err := testutil.NewStoreInstance()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	if err := seeds.SeedInitialData(store); err != nil {
		assert.FailNow(t, err.Error())
	}

	if err := store.UpdateEntity(ctx, entity); err != nil {
		t.Fatal(err)
	}

	keepalive := &corev2.Event{
		Check: &corev2.Check{
			ObjectMeta: corev2.ObjectMeta{
				Name:      "keepalive",
				Namespace: "default",
			},
			Interval: 1,
			Timeout:  5,
			Command:  "command",
		},
		Entity:    entity,
		Timestamp: time.Now().Unix(),
	}

	if _, _, err := store.UpdateEvent(ctx, keepalive); err != nil {
		t.Fatal(err)
	}

	factory := liveness.EtcdFactory(context.Background(), client)

	k, err := New(Config{
		Store:           store,
		EventStore:      store,
		Bus:             bus,
		LivenessFactory: factory,
		BufferSize:      1,
		WorkerCount:     1,
		StoreTimeout:    time.Minute,
	})
	require.NoError(t, err)

	if err := k.Start(); err != nil {
		assert.FailNow(t, err.Error())
	}

	if err := bus.Publish(messaging.TopicKeepalive, keepalive); err != nil {
		assert.FailNow(t, "failed to publish keepalive event")
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

	// Make the entity ephemeral, so that the absence of keepalives leads to its
	// deregistration and no keepalive failure events.
	entity.Deregister = true
	if err := store.UpdateEntity(ctx, entity); err != nil {
		t.Fatal(err)
	}

	// We shouldn't see any keepalive failures after waiting for at least
	// keepalive.Interval + keepalive.Timeout.
	timer := time.NewTimer(time.Duration(keepalive.Check.Interval+keepalive.Check.Timeout) * time.Second)
	select {
	case <-eventChan:
		assert.FailNow(t, "unexpected event received")
	case <-timer.C:
	}

	// Check that the entity has indeed been deregistered
	result, err := store.GetEntityByName(ctx, entity.Name)
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		assert.FailNow(t, "entity did not get deregistered")
	}
}
