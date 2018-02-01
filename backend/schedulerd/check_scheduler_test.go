// +build integration

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
		State: schedulerState,
		Bus:   scheduler.msgBus,
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

func TestCheckSubdueIntervalSuite(t *testing.T) {
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
					End:   "12:30 AM",
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

func TestCheckExecIntervalSuite(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	scheduler := newScheduler(t)

	check := scheduler.check
	request := scheduler.exec.BuildRequest(check)
	assert.NotNil(request)
	assert.NotNil(request.Config)
	assert.NotNil(request.Assets)
	assert.NotEmpty(request.Assets)
	assert.Len(request.Assets, 1)
	assert.NotNil(request.Hooks)
	assert.NotEmpty(request.Hooks)
	assert.Len(request.Hooks, 1)

	check.RuntimeAssets = []string{}
	check.CheckHooks = []types.HookList{}
	request = scheduler.exec.BuildRequest(check)
	assert.NotNil(request)
	assert.NotNil(request.Config)
	assert.Empty(request.Assets)
	assert.Empty(request.Hooks)

	assert.NoError(scheduler.msgBus.Stop())
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
					End:   "12:30 AM",
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

func TestCheckExecCron(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	scheduler := newScheduler(t)

	check := scheduler.check
	check.Cron = "* * * * *"

	request := scheduler.exec.BuildRequest(check)
	assert.NotNil(request)
	assert.NotNil(request.Config)
	assert.NotNil(request.Assets)
	assert.NotEmpty(request.Assets)
	assert.Len(request.Assets, 1)
	assert.NotNil(request.Hooks)
	assert.NotEmpty(request.Hooks)
	assert.Len(request.Hooks, 1)

	check.RuntimeAssets = []string{}
	check.CheckHooks = []types.HookList{}
	request = scheduler.exec.BuildRequest(check)
	assert.NotNil(request)
	assert.NotNil(request.Config)
	assert.Empty(request.Assets)
	assert.Empty(request.Hooks)

	assert.NoError(scheduler.msgBus.Stop())
}

func TestSplayCalculation(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	check := types.FixtureCheckConfig("check1")
	check.ProxyRequests = types.FixtureProxyRequests(true)

	// 10s * 90% / 3 = 3
	check.Interval = 10
	splay, err := calculateSplayInterval(check, 3)
	assert.Equal(float64(3), splay)
	assert.Nil(err)

	// 20s * 50% / 5 = 2
	check.Interval = 20
	check.ProxyRequests.SplayCoverage = 50
	splay, err = calculateSplayInterval(check, 5)
	assert.Equal(float64(2), splay)
	assert.Nil(err)

	// invalid cron string
	check.Cron = "invalid"
	splay, err = calculateSplayInterval(check, 5)
	assert.Equal(float64(0), splay)
	assert.NotNil(err)

	// at most, 60s from current time * 50% / 2 = 15
	// this test will depend on when it is run, but the
	// largest splay calculation will be 15
	check.Cron = "* * * * *"
	splay, err = calculateSplayInterval(check, 2)
	assert.True(splay >= 0 && splay <= 15)
	assert.Nil(err)
}

func TestPublishProxyCheckRequest(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	scheduler := newScheduler(t)

	entity := types.FixtureEntity("entity1")
	check := scheduler.check
	check.Subscriptions = []string{"subscription1"}
	check.ProxyRequests = types.FixtureProxyRequests(true)

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf(
		"%s:%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Organization,
		check.Environment,
	)

	if err := scheduler.msgBus.Subscribe(topic, "TestPublishProxyCheckRequest", c1); err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		assert.NoError(scheduler.msgBus.Unsubscribe(topic, "TestPublishProxyCheckRequest"))
		close(c1)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	go func() {
		select {
		case msg := <-c1:
			res, ok := msg.(*types.CheckRequest)
			assert.True(ok)
			assert.Equal("check1", res.Config.Name)
			assert.Equal("entity1", res.Config.ProxyEntityID)
		}
	}()

	assert.NoError(scheduler.exec.publishProxyCheckRequest(entity, check))
}

func TestPublishProxyCheckRequestsInterval(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	scheduler := newScheduler(t)

	entity1 := types.FixtureEntity("entity1")
	entity2 := types.FixtureEntity("entity2")
	entity3 := types.FixtureEntity("entity3")
	entities := []*types.Entity{entity1, entity2, entity3}
	check := scheduler.check
	check.Subscriptions = []string{"subscription1"}
	check.ProxyRequests = types.FixtureProxyRequests(true)

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf(
		"%s:%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Organization,
		check.Environment,
	)

	if err := scheduler.msgBus.Subscribe(topic, "TestPublishProxyCheckRequestsInterval", c1); err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		assert.NoError(scheduler.msgBus.Unsubscribe(topic, "TestPublishProxyCheckRequestsInterval"))
		close(c1)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	go func() {
		for i := 0; i < len(entities); i++ {
			entityName := fmt.Sprintf("entity%d", i+1)
			select {
			case msg := <-c1:
				res, ok := msg.(*types.CheckRequest)
				assert.True(ok)
				assert.Equal("check1", res.Config.Name)
				assert.Equal(entityName, res.Config.ProxyEntityID)
			}
		}
	}()

	assert.NoError(scheduler.exec.PublishProxyCheckRequests(entities, check))
}

func TestPublishProxyCheckRequestsCron(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	scheduler := newScheduler(t)

	entity1 := types.FixtureEntity("entity1")
	entity2 := types.FixtureEntity("entity2")
	entity3 := types.FixtureEntity("entity3")
	entities := []*types.Entity{entity1, entity2, entity3}
	check := scheduler.check
	check.Subscriptions = []string{"subscription1"}
	check.ProxyRequests = types.FixtureProxyRequests(true)
	check.Cron = "* * * * *"

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf(
		"%s:%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Organization,
		check.Environment,
	)

	if err := scheduler.msgBus.Subscribe(topic, "CheckSchedulerProxySuite", c1); err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		assert.NoError(scheduler.msgBus.Unsubscribe(topic, "CheckSchedulerProxySuite"))
		close(c1)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	go func() {
		for i := 0; i < len(entities); i++ {
			entityName := fmt.Sprintf("entity%d", i+1)
			select {
			case msg := <-c1:
				res, ok := msg.(*types.CheckRequest)
				assert.True(ok)
				assert.Equal("check1", res.Config.Name)
				assert.Equal(entityName, res.Config.ProxyEntityID)
			}
		}
	}()

	assert.NoError(scheduler.exec.PublishProxyCheckRequests(entities, check))
}
