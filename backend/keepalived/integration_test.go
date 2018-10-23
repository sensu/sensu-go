// +build integration

package keepalived

import (
	"testing"
	"time"

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

func TestKeepaliveMonitor(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{
		RingGetter: &mockring.Getter{},
	})
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

	storeInst, err := testutil.NewStoreInstance()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	store := storeInst.GetStore()
	storev2 := storeInst.GetStoreV2()
	if err := seeds.SeedInitialData(store, storev2); err != nil {
		assert.FailNow(t, err.Error())
	}

	mFac := monitor.EtcdFactory(client)

	k, err := New(Config{Store: store, Bus: bus, MonitorFactory: mFac})
	require.NoError(t, err)

	if err := k.Start(); err != nil {
		assert.FailNow(t, err.Error())
	}

	entity := types.FixtureEntity("entity1")

	keepalive := &types.Event{
		Check:     &types.Check{Timeout: 1},
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
