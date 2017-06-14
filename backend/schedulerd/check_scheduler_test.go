package schedulerd

import (
	"fmt"
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

	store := &mockstore.MockStore{}
	check := types.FixtureCheckConfig("check1")
	check.Interval = 1
	check.Subscriptions = []string{"subscription1"}
	store.On("GetCheckConfigByName", "default", "check1").Return(check, nil)

	scheduler := &CheckScheduler{
		MessageBus:  bus,
		Store:       store,
		CheckConfig: check,
		wg:          &sync.WaitGroup{},
	}

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf("%s:%s:subscription1", messaging.TopicSubscriptions, check.Organization)
	assert.NoError(t, bus.Subscribe(topic, "channel1", c1))

	assert.NoError(t, scheduler.Start())
	time.Sleep(1 * time.Second)
	scheduler.Stop()
	assert.NoError(t, bus.Stop())
	close(c1)

	messages := []*types.CheckConfig{}
	for msg := range c1 {
		res, ok := msg.(*types.CheckConfig)
		assert.True(t, ok)
		messages = append(messages, res)
	}
	res := messages[0]
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, "check1", res.Name)
}
