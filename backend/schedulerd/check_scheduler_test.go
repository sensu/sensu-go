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
	"github.com/stretchr/testify/suite"
)

type CheckSchedulerIntervalSuite struct {
	suite.Suite
	check     *types.CheckConfig
	scheduler *CheckScheduler
	msgBus    *messaging.WizardBus
}

func (suite *CheckSchedulerIntervalSuite) SetupTest() {
	suite.check = types.FixtureCheckConfig("check1")
	suite.check.Interval = 1
	suite.msgBus = &messaging.WizardBus{}

	manager := NewStateManager(&mockstore.MockStore{})
	manager.Update(func(state *SchedulerState) {
		state.SetChecks([]*types.CheckConfig{suite.check})
	})

	suite.scheduler = &CheckScheduler{
		CheckName:     suite.check.Name,
		CheckEnv:      suite.check.Environment,
		CheckOrg:      suite.check.Organization,
		CheckInterval: suite.check.Interval,
		CheckCron:     suite.check.Cron,
		LastCronState: suite.check.Cron,
		StateManager:  manager,
		MessageBus:    suite.msgBus,
		WaitGroup:     &sync.WaitGroup{},
	}

	suite.NoError(suite.msgBus.Start())
}

func (suite *CheckSchedulerIntervalSuite) TestStart() {
	// Set interval to smallest valid value
	check := suite.check
	check.Subscriptions = []string{"subscription1"}

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf(
		"%s:%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Organization,
		check.Environment,
	)
	suite.NoError(suite.msgBus.Subscribe(topic, "channel1", c1))

	suite.NoError(suite.scheduler.Start())
	time.Sleep(1 * time.Second)
	suite.NoError(suite.scheduler.Stop())
	suite.NoError(suite.msgBus.Stop())
	close(c1)

	messages := []*types.CheckRequest{}
	for msg := range c1 {
		res, ok := msg.(*types.CheckRequest)
		suite.True(ok)
		messages = append(messages, res)
	}
	res := messages[0]
	suite.Equal(1, len(messages))
	suite.Equal("check1", res.Config.Name)
}

type CheckSubdueIntervalSuite struct {
	suite.Suite
	check     *types.CheckConfig
	scheduler *CheckScheduler
	msgBus    *messaging.WizardBus
}

func (suite *CheckSubdueIntervalSuite) SetupTest() {
	suite.check = types.FixtureCheckConfig("check1")
	suite.check.Interval = 1
	suite.msgBus = &messaging.WizardBus{}

	manager := NewStateManager(&mockstore.MockStore{})
	manager.Update(func(state *SchedulerState) {
		state.SetChecks([]*types.CheckConfig{suite.check})
	})

	suite.scheduler = &CheckScheduler{
		CheckName:     suite.check.Name,
		CheckEnv:      suite.check.Environment,
		CheckOrg:      suite.check.Organization,
		CheckInterval: suite.check.Interval,
		StateManager:  manager,
		MessageBus:    suite.msgBus,
		WaitGroup:     &sync.WaitGroup{},
	}

	suite.NoError(suite.msgBus.Start())
}

func (suite *CheckSubdueIntervalSuite) TestStart() {
	// Set interval to smallest valid value
	check := suite.check
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
	suite.NoError(suite.msgBus.Subscribe(topic, "channel1", c1))

	suite.NoError(suite.scheduler.Start())
	time.Sleep(1 * time.Second)
	suite.NoError(suite.scheduler.Stop())
	suite.NoError(suite.msgBus.Stop())
	close(c1)

	messages := []*types.CheckRequest{}
	for msg := range c1 {
		res, ok := msg.(*types.CheckRequest)
		suite.True(ok)
		messages = append(messages, res)
	}
	// Check should have been subdued at this time, so expect no messages
	suite.Equal(0, len(messages))
}

type TimerIntervalSuite struct {
	suite.Suite
}

func (suite *TimerIntervalSuite) TestSplay() {
	timer := NewIntervalTimer("check1", 10)

	suite.Condition(func() bool { return timer.splay > 0 })

	timer2 := NewIntervalTimer("check1", 10)
	suite.Equal(timer.splay, timer2.splay)
}

func (suite *TimerIntervalSuite) TestInitialOffset() {
	inputs := []uint{1, 10, 60}
	for _, intervalSeconds := range inputs {
		now := time.Now()
		timer := NewIntervalTimer("check1", intervalSeconds)
		nextExecution := timer.calcInitialOffset()
		executionTime := now.Add(nextExecution)

		// We've scheduled it in the future.
		suite.Condition(func() bool { return executionTime.Sub(now) > 0 })
		// The offset is less than the check interval.
		suite.Condition(func() bool { return nextExecution < (time.Duration(intervalSeconds) * time.Second) })
		// The next execution occurs _before_ now + interval.
		suite.Condition(func() bool { return executionTime.Before(now.Add(time.Duration(intervalSeconds) * time.Second)) })
	}
}

func (suite *TimerIntervalSuite) TestStop() {
	timer := NewIntervalTimer("check1", 10)
	timer.Start()

	result := timer.Stop()
	suite.True(result)
}

type CheckExecIntervalSuite struct {
	suite.Suite
	check  *types.CheckConfig
	exec   *CheckExecutor
	msgBus messaging.MessageBus
}

func (suite *CheckExecIntervalSuite) SetupTest() {
	suite.msgBus = &messaging.WizardBus{}
	suite.NoError(suite.msgBus.Start())

	request := types.FixtureCheckRequest("check1")
	asset := request.Assets[0]
	hook := request.Hooks[0]
	suite.check = request.Config

	state := &SchedulerState{}
	state.SetChecks([]*types.CheckConfig{request.Config})
	state.SetAssets([]*types.Asset{&asset})
	state.SetHooks([]*types.HookConfig{&hook})

	suite.exec = &CheckExecutor{
		State: state,
		Bus:   suite.msgBus,
	}
}

func (suite *CheckExecIntervalSuite) AfterTest() {
	suite.NoError(suite.msgBus.Stop())
}

func (suite *CheckExecIntervalSuite) TestBuild() {
	check := suite.check
	request := suite.exec.BuildRequest(check)
	suite.NotNil(request)
	suite.NotNil(request.Config)
	suite.NotNil(request.Assets)
	suite.NotEmpty(request.Assets)
	suite.Len(request.Assets, 1)
	suite.NotNil(request.Hooks)
	suite.NotEmpty(request.Hooks)
	suite.Len(request.Hooks, 1)

	check.RuntimeAssets = []string{}
	check.CheckHooks = []types.HookList{}
	request = suite.exec.BuildRequest(check)
	suite.NotNil(request)
	suite.NotNil(request.Config)
	suite.Empty(request.Assets)
	suite.Empty(request.Hooks)
}

func TestRunExecIntervalSuite(t *testing.T) {
	suite.Run(t, new(TimerIntervalSuite))
	suite.Run(t, new(CheckSchedulerIntervalSuite))
	suite.Run(t, new(CheckExecIntervalSuite))
	suite.Run(t, new(CheckSubdueIntervalSuite))
}

type CheckSchedulerCronSuite struct {
	suite.Suite
	check     *types.CheckConfig
	scheduler *CheckScheduler
	msgBus    *messaging.WizardBus
}

func (suite *CheckSchedulerCronSuite) SetupTest() {
	suite.check = types.FixtureCheckConfig("check1")
	suite.check.Cron = "* * * * *"
	suite.msgBus = &messaging.WizardBus{}

	manager := NewStateManager(&mockstore.MockStore{})
	manager.Update(func(state *SchedulerState) {
		state.SetChecks([]*types.CheckConfig{suite.check})
	})

	suite.scheduler = &CheckScheduler{
		CheckName:     suite.check.Name,
		CheckEnv:      suite.check.Environment,
		CheckOrg:      suite.check.Organization,
		CheckInterval: suite.check.Interval,
		CheckCron:     suite.check.Cron,
		LastCronState: suite.check.Cron,
		StateManager:  manager,
		MessageBus:    suite.msgBus,
		WaitGroup:     &sync.WaitGroup{},
	}

	suite.NoError(suite.msgBus.Start())
}

func (suite *CheckSchedulerCronSuite) TestStart() {
	// Set interval to smallest valid value
	check := suite.check
	check.Subscriptions = []string{"subscription1"}

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf(
		"%s:%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Organization,
		check.Environment,
	)
	suite.NoError(suite.msgBus.Subscribe(topic, "channel1", c1))

	suite.NoError(suite.scheduler.Start())
	time.Sleep(60 * time.Second)
	suite.NoError(suite.scheduler.Stop())
	suite.NoError(suite.msgBus.Stop())
	close(c1)

	messages := []*types.CheckRequest{}
	for msg := range c1 {
		res, ok := msg.(*types.CheckRequest)
		suite.True(ok)
		messages = append(messages, res)
	}
	res := messages[0]
	suite.Equal(1, len(messages))
	suite.Equal("check1", res.Config.Name)
}

type CheckSubdueCronSuite struct {
	suite.Suite
	check     *types.CheckConfig
	scheduler *CheckScheduler
	msgBus    *messaging.WizardBus
}

func (suite *CheckSubdueCronSuite) SetupTest() {
	suite.check = types.FixtureCheckConfig("check1")
	suite.check.Cron = "* * * * *"
	suite.msgBus = &messaging.WizardBus{}

	manager := NewStateManager(&mockstore.MockStore{})
	manager.Update(func(state *SchedulerState) {
		state.SetChecks([]*types.CheckConfig{suite.check})
	})

	suite.scheduler = &CheckScheduler{
		CheckName:     suite.check.Name,
		CheckEnv:      suite.check.Environment,
		CheckOrg:      suite.check.Organization,
		CheckInterval: suite.check.Interval,
		StateManager:  manager,
		MessageBus:    suite.msgBus,
		WaitGroup:     &sync.WaitGroup{},
	}

	suite.NoError(suite.msgBus.Start())
}

func (suite *CheckSubdueCronSuite) TestStart() {
	// Set interval to smallest valid value
	check := suite.check
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
	suite.NoError(suite.msgBus.Subscribe(topic, "channel1", c1))

	suite.NoError(suite.scheduler.Start())
	time.Sleep(60 * time.Second)
	suite.NoError(suite.scheduler.Stop())
	suite.NoError(suite.msgBus.Stop())
	close(c1)

	messages := []*types.CheckRequest{}
	for msg := range c1 {
		res, ok := msg.(*types.CheckRequest)
		suite.True(ok)
		messages = append(messages, res)
	}
	// Check should have been subdued at this time, so expect no messages
	suite.Equal(0, len(messages))
}

type TimerCronSuite struct {
	suite.Suite
}

func (suite *TimerCronSuite) TestStop() {
	timer := NewCronTimer("check1", "* * * * *")
	timer.Start()

	result := timer.Stop()
	suite.True(result)
}

type CheckExecCronSuite struct {
	suite.Suite
	check  *types.CheckConfig
	exec   *CheckExecutor
	msgBus messaging.MessageBus
}

func (suite *CheckExecCronSuite) SetupTest() {
	suite.msgBus = &messaging.WizardBus{}
	suite.NoError(suite.msgBus.Start())

	request := types.FixtureCheckRequest("check1")
	request.Config.Cron = "* * * * *"
	asset := request.Assets[0]
	hook := request.Hooks[0]
	suite.check = request.Config

	state := &SchedulerState{}
	state.SetChecks([]*types.CheckConfig{request.Config})
	state.SetAssets([]*types.Asset{&asset})
	state.SetHooks([]*types.HookConfig{&hook})

	suite.exec = &CheckExecutor{
		State: state,
		Bus:   suite.msgBus,
	}
}

func (suite *CheckExecCronSuite) AfterTest() {
	suite.NoError(suite.msgBus.Stop())
}

func (suite *CheckExecCronSuite) TestBuild() {
	check := suite.check
	request := suite.exec.BuildRequest(check)
	suite.NotNil(request)
	suite.NotNil(request.Config)
	suite.NotNil(request.Assets)
	suite.NotEmpty(request.Assets)
	suite.Len(request.Assets, 1)
	suite.NotNil(request.Hooks)
	suite.NotEmpty(request.Hooks)
	suite.Len(request.Hooks, 1)

	check.RuntimeAssets = []string{}
	check.CheckHooks = []types.HookList{}
	request = suite.exec.BuildRequest(check)
	suite.NotNil(request)
	suite.NotNil(request.Config)
	suite.Empty(request.Assets)
	suite.Empty(request.Hooks)
}

func TestRunExecCronSuite(t *testing.T) {
	suite.Run(t, new(TimerCronSuite))
	suite.Run(t, new(CheckSchedulerCronSuite))
	suite.Run(t, new(CheckExecCronSuite))
	suite.Run(t, new(CheckSubdueCronSuite))
}

type CheckSchedulerProxySuite struct {
	suite.Suite
	check  *types.CheckConfig
	exec   *CheckExecutor
	msgBus *messaging.WizardBus
}

func (suite *CheckSchedulerProxySuite) SetupTest() {
	suite.msgBus = &messaging.WizardBus{}
	suite.NoError(suite.msgBus.Start())

	request := types.FixtureCheckRequest("check1")
	asset := request.Assets[0]
	hook := request.Hooks[0]
	suite.check = request.Config
	suite.check.Interval = 10

	state := &SchedulerState{}
	state.SetChecks([]*types.CheckConfig{request.Config})
	state.SetAssets([]*types.Asset{&asset})
	state.SetHooks([]*types.HookConfig{&hook})

	suite.exec = &CheckExecutor{
		State: state,
		Bus:   suite.msgBus,
	}
}

func (suite *CheckSchedulerProxySuite) TestSplayCalculation() {
	check := types.FixtureCheckConfig("check1")
	check.ProxyRequests = types.FixtureProxyRequests(true)

	// 10s * 90% / 3 = 3
	check.Interval = 10
	splay, err := calculateSplayInterval(check, 3)
	suite.Equal(float64(3), splay)
	suite.Nil(err)

	// 20s * 50% / 5 = 2
	check.Interval = 20
	check.ProxyRequests.SplayCoverage = 50
	splay, err = calculateSplayInterval(check, 5)
	suite.Equal(float64(2), splay)
	suite.Nil(err)

	// invalid cron string
	check.Cron = "invalid"
	splay, err = calculateSplayInterval(check, 5)
	suite.Equal(float64(0), splay)
	suite.NotNil(err)

	// at most, 60s from current time * 50% / 2 = 15
	// this test will depend on when it is run, but the
	// largest splay calculation will be 15
	check.Cron = "* * * * *"
	splay, err = calculateSplayInterval(check, 2)
	suite.True(splay >= 0 && splay <= 15)
	suite.Nil(err)
}

func (suite *CheckSchedulerProxySuite) TestPublishProxyCheckRequest() {
	entity := types.FixtureEntity("entity1")
	check := suite.check
	check.Subscriptions = []string{"subscription1"}
	check.ProxyRequests = types.FixtureProxyRequests(true)

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf(
		"%s:%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Organization,
		check.Environment,
	)
	suite.NoError(suite.msgBus.Subscribe(topic, "channel1", c1))

	suite.NoError(suite.exec.publishProxyCheckRequest(entity, check))
	suite.NoError(suite.msgBus.Stop())
	close(c1)

	messages := []*types.CheckRequest{}
	for msg := range c1 {
		res, ok := msg.(*types.CheckRequest)
		suite.True(ok)
		messages = append(messages, res)
	}
	res := messages[0]
	suite.Equal(1, len(messages))
	suite.Equal("check1", res.Config.Name)
	suite.Equal("entity1", res.Config.ProxyEntityID)
}

func (suite *CheckSchedulerProxySuite) TestPublishProxyCheckRequestsInterval() {
	entity1 := types.FixtureEntity("entity1")
	entity2 := types.FixtureEntity("entity2")
	entity3 := types.FixtureEntity("entity3")
	entities := []*types.Entity{entity1, entity2, entity3}
	check := suite.check
	check.Subscriptions = []string{"subscription1"}
	check.ProxyRequests = types.FixtureProxyRequests(true)

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf(
		"%s:%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Organization,
		check.Environment,
	)
	suite.NoError(suite.msgBus.Subscribe(topic, "channel1", c1))

	go func() {
		for i := 0; i < len(entities); i++ {
			entityName := fmt.Sprintf("entity%d", i+1)
			select {
			case msg := <-c1:
				res, ok := msg.(*types.CheckRequest)
				suite.True(ok)
				suite.Equal("check1", res.Config.Name)
				suite.Equal(entityName, res.Config.ProxyEntityID)
			}
		}
	}()
	suite.NoError(suite.exec.PublishProxyCheckRequests(entities, check))
	suite.NoError(suite.msgBus.Stop())
	close(c1)
}

func (suite *CheckSchedulerProxySuite) TestPublishProxyCheckRequestsCron() {
	entity1 := types.FixtureEntity("entity1")
	entity2 := types.FixtureEntity("entity2")
	entity3 := types.FixtureEntity("entity3")
	entities := []*types.Entity{entity1, entity2, entity3}
	check := suite.check
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
	suite.NoError(suite.msgBus.Subscribe(topic, "channel1", c1))

	go func() {
		for i := 0; i < len(entities); i++ {
			entityName := fmt.Sprintf("entity%d", i+1)
			select {
			case msg := <-c1:
				res, ok := msg.(*types.CheckRequest)
				suite.True(ok)
				suite.Equal("check1", res.Config.Name)
				suite.Equal(entityName, res.Config.ProxyEntityID)
			}
		}
	}()
	suite.NoError(suite.exec.PublishProxyCheckRequests(entities, check))
	suite.NoError(suite.msgBus.Stop())
	close(c1)
}

func TestRunExecProxySuite(t *testing.T) {
	suite.Run(t, new(CheckSchedulerProxySuite))
}
