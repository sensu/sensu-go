// +build integration,race

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

type TestCheckScheduler struct {
	check     *types.CheckConfig
	exec      *CheckExecutor
	msgBus    *messaging.WizardBus
	scheduler *CheckScheduler
}

func newScheduler(t *testing.T) *TestCheckScheduler {
	t.Helper()

	assert := assert.New(t)

	scheduler := &TestCheckScheduler{}

	request := types.FixtureCheckRequest("check1")
	asset := request.Assets[0]
	hook := request.Hooks[0]
	scheduler.check = request.Config
	scheduler.check.Interval = 1

	scheduler.msgBus = &messaging.WizardBus{}
	schedulerState := &SchedulerState{}

	manager := NewStateManager(&mockstore.MockStore{})
	manager.Update(func(state *SchedulerState) {
		state.SetChecks([]*types.CheckConfig{scheduler.check})
		state.SetAssets([]*types.Asset{&asset})
		state.SetHooks([]*types.HookConfig{&hook})
		schedulerState = state
	})

	scheduler.scheduler = &CheckScheduler{
		CheckName:     scheduler.check.Name,
		CheckEnv:      scheduler.check.Environment,
		CheckOrg:      scheduler.check.Organization,
		CheckInterval: scheduler.check.Interval,
		CheckCron:     scheduler.check.Cron,
		LastCronState: scheduler.check.Cron,
		StateManager:  manager,
		MessageBus:    scheduler.msgBus,
		WaitGroup:     &sync.WaitGroup{},
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

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf(
		"%s:%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Organization,
		check.Environment,
	)

	if err := scheduler.msgBus.Subscribe(topic, "CheckSchedulerIntervalSuite", c1); err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		assert.NoError(scheduler.msgBus.Unsubscribe(topic, "CheckSchedulerIntervalSuite"))
		close(c1)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	go func() {
		select {
		case msg := <-c1:
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

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf(
		"%s:%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Organization,
		check.Environment,
	)

	if err := scheduler.msgBus.Subscribe(topic, "CheckSubdueIntervalSuite", c1); err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		assert.NoError(scheduler.msgBus.Unsubscribe(topic, "CheckSubdueIntervalSuite"))
		close(c1)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	assert.NoError(scheduler.scheduler.Start())
	time.Sleep(1 * time.Second)
	assert.NoError(scheduler.scheduler.Stop())

	// We should have no element in our channel
	assert.Equal(0, len(c1))
}

func TestCheckSchedulerCron(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	scheduler := newScheduler(t)

	// Set interval to smallest valid value
	check := scheduler.check
	check.Subscriptions = []string{"subscription1"}

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf(
		"%s:%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Organization,
		check.Environment,
	)

	if err := scheduler.msgBus.Subscribe(topic, "CheckSchedulerCronSuite", c1); err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		assert.NoError(scheduler.msgBus.Unsubscribe(topic, "CheckSchedulerCronSuite"))
		close(c1)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	go func() {
		select {
		case msg := <-c1:
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

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf(
		"%s:%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Organization,
		check.Environment,
	)

	if err := scheduler.msgBus.Subscribe(topic, "CheckSubdueCronSuite", c1); err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		assert.NoError(scheduler.msgBus.Unsubscribe(topic, "CheckSubdueCronSuite"))
		close(c1)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	assert.NoError(scheduler.scheduler.Start())
	time.Sleep(60 * time.Second)
	assert.NoError(scheduler.scheduler.Stop())

	// We should have no element in our channel
	assert.Equal(0, len(c1))
}
