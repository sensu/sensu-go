package schedulerd

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/suite"
)

type CheckSchedulerSuite struct {
	suite.Suite
	check     *types.CheckConfig
	scheduler *CheckScheduler
	msgBus    *messaging.WizardBus
}

func (suite *CheckSchedulerSuite) SetupTest() {
	suite.check = types.FixtureCheckConfig("check1")
	suite.msgBus = &messaging.WizardBus{}

	manager := NewStateManager()
	manager.Update(func(state *SchedulerState) {
		state.SetChecks([]*types.CheckConfig{suite.check})
	})

	suite.scheduler = &CheckScheduler{
		CheckName:    suite.check.Name,
		CheckOrg:     suite.check.Organization,
		StateManager: manager,
		MessageBus:   suite.msgBus,
		WaitGroup:    &sync.WaitGroup{},
	}

	suite.msgBus.Start()
}

func (suite *CheckSchedulerSuite) TestStart() {
	// Set interval to smallest valid value
	check := suite.check
	check.Interval = 1
	check.Subscriptions = []string{"subscription1"}

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf("%s:%s:subscription1", messaging.TopicSubscriptions, check.Organization)
	suite.NoError(suite.msgBus.Subscribe(topic, "channel1", c1))

	suite.NoError(suite.scheduler.Start(1))
	time.Sleep(1 * time.Second)
	suite.scheduler.Stop()
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

type TimerSuite struct {
	suite.Suite
}

func (suite *TimerSuite) TestSplay() {
	timer := NewCheckTimer("check1", 10)

	suite.Condition(func() bool { return timer.splay > 0 })

	timer2 := NewCheckTimer("check1", 10)
	suite.Equal(timer.splay, timer2.splay)
}

func (suite *TimerSuite) TestInitialOffset() {
	inputs := []int{1, 10, 60}
	for _, intervalSeconds := range inputs {
		now := time.Now()
		timer := NewCheckTimer("check1", intervalSeconds)
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

func (suite *TimerSuite) TestStop() {
	timer := NewCheckTimer("check1", 10)
	timer.Start()

	result := timer.Stop()
	suite.True(result)
}

type CheckExecSuite struct {
	suite.Suite
	check  *types.CheckConfig
	exec   *CheckExecutor
	msgBus messaging.MessageBus
}

func (suite *CheckExecSuite) SetupTest() {
	suite.msgBus = &messaging.WizardBus{}
	suite.msgBus.Start()

	request := types.FixtureCheckRequest("check1")
	asset := request.ExpandedAssets[0]
	suite.check = request.Config

	state := &SchedulerState{}
	state.SetChecks([]*types.CheckConfig{request.Config})
	state.SetAssets([]*types.Asset{&asset})

	suite.exec = &CheckExecutor{
		State: state,
		Bus:   suite.msgBus,
	}
}

func (suite *CheckExecSuite) AfterTest() {
	suite.msgBus.Stop()
}

func (suite *CheckExecSuite) TestBuild() {
	check := suite.check
	request := suite.exec.BuildRequest(check)
	suite.NotNil(request)
	suite.NotNil(request.Config)
	suite.NotNil(request.ExpandedAssets)
	suite.NotEmpty(request.ExpandedAssets)
	suite.Len(request.ExpandedAssets, 1)

	check.RuntimeAssets = []string{}
	request = suite.exec.BuildRequest(check)
	suite.NotNil(request)
	suite.NotNil(request.Config)
	suite.Empty(request.ExpandedAssets)
}

func TestRunExecSuite(t *testing.T) {
	suite.Run(t, new(TimerSuite))
	suite.Run(t, new(CheckSchedulerSuite))
	suite.Run(t, new(CheckExecSuite))
}
