// +build integration,race

package schedulerd

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockring"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestCheckScheduler struct {
	check        *types.CheckConfig
	exec         *CheckExecutor
	msgBus       *messaging.WizardBus
	scheduler    *CheckScheduler
	channel      chan interface{}
	subscription messaging.Subscription
}

func (tcs *TestCheckScheduler) Receiver() chan<- interface{} {
	return tcs.channel
}

func newScheduler(t *testing.T) *TestCheckScheduler {
	t.Helper()

	assert := assert.New(t)

	scheduler := &TestCheckScheduler{}
	scheduler.channel = make(chan interface{}, 2)

	request := types.FixtureCheckRequest("check1")
	asset := request.Assets[0]
	hook := request.Hooks[0]
	scheduler.check = request.Config
	scheduler.check.Interval = 1

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{
		RingGetter: &mockring.Getter{},
	})
	require.NoError(t, err)
	scheduler.msgBus = bus
	schedulerState := &SchedulerState{}

	manager := NewStateManager(&mockstore.MockStore{})
	manager.Update(func(state *SchedulerState) {
		state.SetChecks([]*types.CheckConfig{scheduler.check})
		state.SetAssets([]*types.Asset{&asset})
		state.SetHooks([]*types.HookConfig{&hook})
		schedulerState = state
	})

	scheduler.scheduler = &CheckScheduler{
		checkName:     scheduler.check.Name,
		checkEnv:      scheduler.check.Environment,
		checkOrg:      scheduler.check.Organization,
		checkInterval: scheduler.check.Interval,
		checkCron:     scheduler.check.Cron,
		lastCronState: scheduler.check.Cron,
		stateManager:  manager,
		bus:           scheduler.msgBus,
		wg:            &sync.WaitGroup{},
	}

	assert.NoError(scheduler.msgBus.Start())

	scheduler.exec = &CheckExecutor{
		state: schedulerState,
		bus:   scheduler.msgBus,
	}

	return scheduler
}

func TestCheckSchedulerInterval(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	scheduler := newScheduler(t)

	// Set interval to smallest valid value
	check := scheduler.check
	check.Subscriptions = []string{"subscription1"}

	topic := messaging.SubscriptionTopic(check.Organization, check.Environment, "subscription1")

	sub, err := scheduler.msgBus.Subscribe(topic, "scheduler", scheduler)
	if err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		sub.Cancel()
		close(scheduler.channel)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	go func() {
		select {
		case msg := <-scheduler.channel:
			res, ok := msg.(*types.CheckRequest)
			assert.True(ok)
			assert.Equal("check1", res.Config.Name)
		}
	}()

	assert.NoError(scheduler.scheduler.Start())
	time.Sleep(5 * time.Second)
	assert.NoError(scheduler.scheduler.Stop())
}

func TestCheckSubdueInterval(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	scheduler := newScheduler(t)

	// Set interval to smallest valid value
	check := scheduler.check
	check.Subscriptions = []string{"subscription1"}
	check.Subdue = &types.TimeWindowWhen{
		Days: types.TimeWindowDays{
			All: []*types.TimeWindowTimeRange{
				{
					Begin: "1:00 AM",
					End:   "11:00 PM",
				},
				{
					Begin: "10:00 PM",
					End:   "2:00 AM",
				},
			},
		},
	}

	topic := fmt.Sprintf(
		"%s:%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Organization,
		check.Environment,
	)

	subscription, err := scheduler.msgBus.Subscribe(topic, "scheduler", scheduler)
	if err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		subscription.Cancel()
		close(scheduler.channel)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	assert.NoError(scheduler.scheduler.Start())
	time.Sleep(1 * time.Second)
	assert.NoError(scheduler.scheduler.Stop())

	// We should have no element in our channel
	assert.Equal(0, len(scheduler.channel))
}

func TestCheckSchedulerCron(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	scheduler := newScheduler(t)

	// Set interval to smallest valid value
	check := scheduler.check
	check.Subscriptions = []string{"subscription1"}

	topic := fmt.Sprintf(
		"%s:%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Organization,
		check.Environment,
	)

	subscription, err := scheduler.msgBus.Subscribe(topic, "scheduler", scheduler)
	if err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		subscription.Cancel()
		close(scheduler.channel)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	go func() {
		select {
		case msg := <-scheduler.channel:
			res, ok := msg.(*types.CheckRequest)
			assert.True(ok)
			assert.Equal("check1", res.Config.Name)
		}
	}()

	assert.NoError(scheduler.scheduler.Start())
	time.Sleep(60 * time.Second)
	assert.NoError(scheduler.scheduler.Stop())
}

func TestCheckSubdueCron(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	scheduler := newScheduler(t)

	// Set interval to smallest valid value
	check := scheduler.check
	check.Cron = "* * * * *"
	check.Subscriptions = []string{"subscription1"}
	check.Subdue = &types.TimeWindowWhen{
		Days: types.TimeWindowDays{
			All: []*types.TimeWindowTimeRange{
				{
					Begin: "1:00 AM",
					End:   "11:00 PM",
				},
				{
					Begin: "10:00 PM",
					End:   "2:00 AM",
				},
			},
		},
	}

	topic := fmt.Sprintf(
		"%s:%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Organization,
		check.Environment,
	)

	subscription, err := scheduler.msgBus.Subscribe(topic, "scheduler", scheduler)
	if err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		subscription.Cancel()
		close(scheduler.channel)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	assert.NoError(scheduler.scheduler.Start())
	time.Sleep(60 * time.Second)
	assert.NoError(scheduler.scheduler.Stop())

	// We should have no element in our channel
	assert.Equal(0, len(scheduler.channel))
}
