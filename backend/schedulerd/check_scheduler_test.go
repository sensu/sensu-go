package schedulerd

import (
	"sync"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestCheckScheduler(t *testing.T) {
	bus := &messaging.WizardBus{}
	assert.NoError(t, bus.Start())

	st := &mockstore.MockStore{}
	check := types.FixtureCheck("check1")
	check.Interval = 1
	check.Subscriptions = []string{"subscription1"}
	st.On("GetCheckByName", "check1").Return(check, nil)

	scheduler := &CheckScheduler{
		MessageBus: bus,
		Store:      st,
		Check:      check,
		wg:         &sync.WaitGroup{},
	}

	c1 := make(chan interface{}, 10)
	assert.NoError(t, bus.Subscribe("subscription1", "channel1", c1))

	assert.NoError(t, scheduler.Start())
	time.Sleep(1 * time.Second)
	scheduler.Stop()
	assert.NoError(t, bus.Stop())

	messages := []*types.Event{}
	for msg := range c1 {
		evt, ok := msg.(*types.Event)
		assert.True(t, ok)
		messages = append(messages, evt)
	}
	evt := messages[0]
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, "check1", evt.Check.Name)
}
