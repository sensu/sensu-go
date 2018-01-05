package keepalived

import (
	"flag"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/seeds"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var integration = flag.Bool("integration", false, "run integration tests")

// send keepalive, expect pass event, allow timer to expire, expect fail event
// mock the message bus, use a store

func TestKeepaliveMonitor(t *testing.T) {
	bus := &mockbus.MockBus{}

	okStatusCount := 0
	warnStatusCount := 0

	bus.On("Publish", messaging.TopicEventRaw, mock.MatchedBy(func(msg interface{}) bool {
		event := msg.(*types.Event)
		if event.Check.Config.Name == "keepalive" && event.Check.Status == 0 {
			okStatusCount++
			return true
		}
		return false
	})).Return(nil)

	bus.On("Publish", messaging.TopicEventRaw, mock.MatchedBy(func(msg interface{}) bool {
		event := msg.(*types.Event)
		if event.Check.Config.Name == "keepalive" && event.Check.Status == 1 {
			warnStatusCount++
			return true
		}
		return false
	})).Return(nil)

	store, err := testutil.NewStoreInstance()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	if err := seeds.SeedInitialData(store); err != nil {
		assert.FailNow(t, err.Error())
	}

	entity := types.FixtureEntity("entity1")
	entity.KeepaliveTimeout = 1

	monitor := KeepaliveMonitor{
		Entity:     entity,
		MessageBus: bus,
		Store:      store,
		EventCreator: &MessageBusEventCreator{
			MessageBus: bus,
		},
	}

	monitor.Start()

	keepalive := &types.Event{
		Entity:    entity,
		Timestamp: time.Now().Unix(),
	}

	if err := monitor.Update(keepalive); err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, 1, okStatusCount)
	assert.Equal(t, 0, warnStatusCount)

	time.Sleep(1 * time.Second)

	assert.Equal(t, 1, okStatusCount)
	assert.Equal(t, 1, warnStatusCount)
}
