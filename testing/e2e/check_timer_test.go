package e2e

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/schedulerd"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/suite"
)

type CheckSchedulerIntervalSuite struct {
	suite.Suite
	check     *types.CheckConfig
	scheduler *schedulerd.CheckScheduler
	msgBus    *messaging.WizardBus
}

func (suite *CheckSchedulerIntervalSuite) SetupTest() {
	suite.check = types.FixtureCheckConfig("check1")
	suite.check.Interval = 1
	suite.msgBus = &messaging.WizardBus{}

	manager := schedulerd.NewStateManager(&mockstore.MockStore{})
	manager.Update(func(state *schedulerd.SchedulerState) {
		state.SetChecks([]*types.CheckConfig{suite.check})
	})

	suite.scheduler = &schedulerd.CheckScheduler{
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
	scheduler *schedulerd.CheckScheduler
	msgBus    *messaging.WizardBus
}

func (suite *CheckSubdueIntervalSuite) SetupTest() {
	suite.check = types.FixtureCheckConfig("check1")
	suite.check.Interval = 1
	suite.msgBus = &messaging.WizardBus{}

	manager := schedulerd.NewStateManager(&mockstore.MockStore{})
	manager.Update(func(state *schedulerd.SchedulerState) {
		state.SetChecks([]*types.CheckConfig{suite.check})
	})

	suite.scheduler = &schedulerd.CheckScheduler{
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

func (suite *TimerIntervalSuite) TestStop() {
	timer := schedulerd.NewIntervalTimer("check1", 10)
	timer.Start()

	result := timer.Stop()
	suite.True(result)
}

type CheckExecIntervalSuite struct {
	suite.Suite
	check  *types.CheckConfig
	exec   *schedulerd.CheckExecutor
	msgBus messaging.MessageBus
}

func (suite *CheckExecIntervalSuite) SetupTest() {
	suite.msgBus = &messaging.WizardBus{}
	suite.NoError(suite.msgBus.Start())

	request := types.FixtureCheckRequest("check1")
	asset := request.Assets[0]
	hook := request.Hooks[0]
	suite.check = request.Config

	state := &schedulerd.SchedulerState{}
	state.SetChecks([]*types.CheckConfig{request.Config})
	state.SetAssets([]*types.Asset{&asset})
	state.SetHooks([]*types.HookConfig{&hook})

	suite.exec = &schedulerd.CheckExecutor{
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
	scheduler *schedulerd.CheckScheduler
	msgBus    *messaging.WizardBus
}

func (suite *CheckSchedulerCronSuite) SetupTest() {
	suite.check = types.FixtureCheckConfig("check1")
	suite.check.Cron = "* * * * *"
	suite.msgBus = &messaging.WizardBus{}

	manager := schedulerd.NewStateManager(&mockstore.MockStore{})
	manager.Update(func(state *schedulerd.SchedulerState) {
		state.SetChecks([]*types.CheckConfig{suite.check})
	})

	suite.scheduler = &schedulerd.CheckScheduler{
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
	scheduler *schedulerd.CheckScheduler
	msgBus    *messaging.WizardBus
}

func (suite *CheckSubdueCronSuite) SetupTest() {
	suite.check = types.FixtureCheckConfig("check1")
	suite.check.Cron = "* * * * *"
	suite.msgBus = &messaging.WizardBus{}

	manager := schedulerd.NewStateManager(&mockstore.MockStore{})
	manager.Update(func(state *schedulerd.SchedulerState) {
		state.SetChecks([]*types.CheckConfig{suite.check})
	})

	suite.scheduler = &schedulerd.CheckScheduler{
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
	timer := schedulerd.NewCronTimer("check1", "* * * * *")
	timer.Start()

	result := timer.Stop()
	suite.True(result)
}

type CheckExecCronSuite struct {
	suite.Suite
	check  *types.CheckConfig
	exec   *schedulerd.CheckExecutor
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

	state := &schedulerd.SchedulerState{}
	state.SetChecks([]*types.CheckConfig{request.Config})
	state.SetAssets([]*types.Asset{&asset})
	state.SetHooks([]*types.HookConfig{&hook})

	suite.exec = &schedulerd.CheckExecutor{
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
