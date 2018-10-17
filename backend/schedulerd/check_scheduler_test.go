package schedulerd

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockring"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func newScheduler(t *testing.T, ctx context.Context) *TestCheckScheduler {
	t.Helper()

	assert := assert.New(t)

	scheduler := &TestCheckScheduler{}
	scheduler.channel = make(chan interface{}, 2)

	request := types.FixtureCheckRequest("check1")
	asset := request.Assets[0]
	hook := request.Hooks[0]
	scheduler.check = request.Config
	scheduler.check.Interval = 1
	store := &mockstore.MockStore{}
	store.On("GetAssets", mock.Anything).Return([]*types.Asset{&asset}, nil)
	store.On("GetHookConfigs", mock.Anything).Return([]*types.HookConfig{&hook}, nil)
	store.On("GetCheckConfigByName", mock.Anything, mock.Anything).Return(scheduler.check, nil)

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{
		RingGetter: &mockring.Getter{},
	})
	require.NoError(t, err)
	scheduler.msgBus = bus

	scheduler.scheduler = NewCheckScheduler(store, scheduler.msgBus, scheduler.check, ctx)

	assert.NoError(scheduler.msgBus.Start())

	roundRobin := newRoundRobinScheduler(ctx, scheduler.msgBus)
	scheduler.exec = NewCheckExecutor(scheduler.msgBus, roundRobin, "default", store)

	return scheduler
}

func TestCheckSchedulerInterval(t *testing.T) {
	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newScheduler(t, ctx)

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
		select {
		case msg := <-scheduler.channel:
			res, ok := msg.(*types.CheckRequest)
			assert.True(ok)
			assert.Equal("check1", res.Config.Name)
			wg.Done()
		}
	}()

	assert.NoError(scheduler.scheduler.Start())
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
	scheduler := newScheduler(t, ctx)

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

	assert.NoError(scheduler.scheduler.Start())
	mockTime.Set(mockTime.Now().Add(2 * time.Second))
	assert.NoError(scheduler.scheduler.Stop())

	// We should have no element in our channel
	assert.Equal(0, len(scheduler.channel))
}

func TestCheckSchedulerCron(t *testing.T) {
	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newScheduler(t, ctx)

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
		select {
		case msg := <-scheduler.channel:
			res, ok := msg.(*types.CheckRequest)
			assert.True(ok)
			assert.Equal("check1", res.Config.Name)
			wg.Done()
		}
	}()

	assert.NoError(scheduler.scheduler.Start())
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
	scheduler := newScheduler(t, ctx)

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

	assert.NoError(scheduler.scheduler.Start())
	mockTime.Set(mockTime.Now().Add(10 * time.Second))
	assert.NoError(scheduler.scheduler.Stop())

	// We should have no element in our channel
	assert.Equal(0, len(scheduler.channel))
}
