package schedulerd

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/store"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type TestIntervalScheduler struct {
	check     *types.CheckConfig
	exec      Executor
	msgBus    *messaging.WizardBus
	scheduler *IntervalScheduler
	channel   chan interface{}
}

func (tcs *TestIntervalScheduler) Receiver() chan<- interface{} {
	return tcs.channel
}

type TestCronScheduler struct {
	check     *types.CheckConfig
	exec      Executor
	msgBus    *messaging.WizardBus
	scheduler *CronScheduler
	channel   chan interface{}
}

func (tcs *TestCronScheduler) Receiver() chan<- interface{} {
	return tcs.channel
}

func newIntervalScheduler(t *testing.T, ctx context.Context, executor string) *TestIntervalScheduler {
	t.Helper()

	assert := assert.New(t)

	scheduler := &TestIntervalScheduler{}
	scheduler.channel = make(chan interface{}, 2)

	request := types.FixtureCheckRequest("check1")
	asset := request.Assets[0]
	hook := request.Hooks[0]
	scheduler.check = request.Config
	scheduler.check.Interval = 1
	s := &mockstore.MockStore{}
	s.On("GetAssets", mock.Anything, &store.SelectionPredicate{}).Return([]*types.Asset{&asset}, nil)
	s.On("GetHookConfigs", mock.Anything, &store.SelectionPredicate{}).Return([]*types.HookConfig{&hook}, nil)
	s.On("GetCheckConfigByName", mock.Anything, mock.Anything).Return(scheduler.check, nil)

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	scheduler.msgBus = bus

	scheduler.scheduler = NewIntervalScheduler(ctx, s, scheduler.msgBus, scheduler.check, &EntityCache{})

	assert.NoError(scheduler.msgBus.Start())

	switch executor {
	case "adhoc":
		scheduler.exec = NewAdhocRequestExecutor(ctx, s, &queue.Memory{}, scheduler.msgBus, &EntityCache{})
	default:
		scheduler.exec = NewCheckExecutor(scheduler.msgBus, "default", s, &EntityCache{})
	}

	return scheduler
}

func newCronScheduler(t *testing.T, ctx context.Context, executor string) *TestCronScheduler {
	t.Helper()

	assert := assert.New(t)

	scheduler := &TestCronScheduler{}
	scheduler.channel = make(chan interface{}, 2)

	request := types.FixtureCheckRequest("check1")
	asset := request.Assets[0]
	hook := request.Hooks[0]
	scheduler.check = request.Config
	scheduler.check.Interval = 1
	scheduler.check.Cron = "* * * * *"
	s := &mockstore.MockStore{}
	s.On("GetAssets", mock.Anything, &store.SelectionPredicate{}).Return([]*types.Asset{&asset}, nil)
	s.On("GetHookConfigs", mock.Anything, &store.SelectionPredicate{}).Return([]*types.HookConfig{&hook}, nil)
	s.On("GetCheckConfigByName", mock.Anything, mock.Anything).Return(scheduler.check, nil)

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	scheduler.msgBus = bus

	scheduler.scheduler = NewCronScheduler(ctx, s, scheduler.msgBus, scheduler.check, &EntityCache{})

	assert.NoError(scheduler.msgBus.Start())

	switch executor {
	case "adhoc":
		scheduler.exec = NewAdhocRequestExecutor(ctx, s, &queue.Memory{}, scheduler.msgBus, &EntityCache{})
	default:
		scheduler.exec = NewCheckExecutor(scheduler.msgBus, "default", s, &EntityCache{})
	}

	return scheduler
}

func TestIntervalScheduling(t *testing.T) {
	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newIntervalScheduler(t, ctx, "check")

	// Set interval to smallest valid value
	check := scheduler.check
	check.Subscriptions = []string{"subscription1"}

	topic := messaging.SubscriptionTopic(check.Namespace, "subscription1")

	sub, err := scheduler.msgBus.Subscribe(topic, "scheduler", scheduler)
	if err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		sub.Cancel()
		assert.NoError(scheduler.msgBus.Stop())
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		msg := <-scheduler.channel
		res, ok := msg.(*types.CheckRequest)
		assert.True(ok)
		assert.Equal("check1", res.Config.Name)
		wg.Done()
	}()

	scheduler.scheduler.Start()
	mockTime.Start()
	wg.Wait()
	mockTime.Stop()
	assert.NoError(scheduler.scheduler.Stop())
}

func TestCheckSubdueInterval(t *testing.T) {
	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newIntervalScheduler(t, ctx, "check")

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
		"%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Namespace,
	)

	subscription, err := scheduler.msgBus.Subscribe(topic, "scheduler", scheduler)
	if err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		subscription.Cancel()
		assert.NoError(scheduler.msgBus.Stop())
	}()

	scheduler.scheduler.Start()
	mockTime.Set(mockTime.Now().Add(2 * time.Second))
	assert.NoError(scheduler.scheduler.Stop())

	// We should have no element in our channel
	assert.Equal(0, len(scheduler.channel))
}

func TestCronScheduling(t *testing.T) {
	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newCronScheduler(t, ctx, "check")

	// Set interval to smallest valid value
	check := scheduler.check
	check.Cron = "* * * * *"
	check.Subscriptions = []string{"subscription1"}

	topic := fmt.Sprintf(
		"%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Namespace,
	)

	subscription, err := scheduler.msgBus.Subscribe(topic, "scheduler", scheduler)
	if err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		subscription.Cancel()
		assert.NoError(scheduler.msgBus.Stop())
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		msg := <-scheduler.channel
		res, ok := msg.(*types.CheckRequest)
		assert.True(ok)
		assert.Equal("check1", res.Config.Name)
		wg.Done()
	}()

	scheduler.scheduler.Start()
	mockTime.Start()
	wg.Wait()
	mockTime.Stop()
	assert.NoError(scheduler.scheduler.Stop())
}

func TestCheckSubdueCron(t *testing.T) {
	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newCronScheduler(t, ctx, "check")

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
		"%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Namespace,
	)

	subscription, err := scheduler.msgBus.Subscribe(topic, "scheduler", scheduler)
	if err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		subscription.Cancel()
		assert.NoError(scheduler.msgBus.Stop())
	}()

	scheduler.scheduler.Start()
	mockTime.Set(mockTime.Now().Add(10 * time.Second))
	assert.NoError(scheduler.scheduler.Stop())

	// We should have no element in our channel
	assert.Equal(0, len(scheduler.channel))
}
