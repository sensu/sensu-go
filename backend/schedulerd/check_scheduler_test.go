package schedulerd

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	cachev2 "github.com/sensu/sensu-go/backend/store/cache/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type TestIntervalScheduler struct {
	check     *corev2.CheckConfig
	exec      Executor
	msgBus    *messaging.WizardBus
	scheduler *IntervalScheduler
	channel   chan interface{}
}

func (tcs *TestIntervalScheduler) Receiver() chan<- interface{} {
	return tcs.channel
}

type TestCronScheduler struct {
	check     *corev2.CheckConfig
	exec      Executor
	msgBus    *messaging.WizardBus
	scheduler *CronScheduler
	channel   chan interface{}
}

func (tcs *TestCronScheduler) Receiver() chan<- interface{} {
	return tcs.channel
}

type mockEventReceiver struct {
	mock.Mock
}

func (m *mockEventReceiver) GenerateBackendEvent(component string, status uint32, output string) error {
	args := m.Called(component, status, output)
	return args.Error(0)
}

func newIntervalScheduler(ctx context.Context, t *testing.T, executor string) *TestIntervalScheduler {
	t.Helper()

	assert := assert.New(t)

	scheduler := &TestIntervalScheduler{}
	scheduler.channel = make(chan interface{}, 2)

	request := corev2.FixtureCheckRequest("check1")
	asset := request.Assets[0]
	hook := request.Hooks[0]
	scheduler.check = request.Config
	scheduler.check.Interval = 1
	s := &mockstore.MockStore{}
	s.On("GetAssets", mock.Anything, &store.SelectionPredicate{}).Return([]*corev2.Asset{&asset}, nil)
	s.On("GetHookConfigs", mock.Anything, &store.SelectionPredicate{}).Return([]*corev2.HookConfig{&hook}, nil)
	s.On("GetCheckConfigByName", mock.Anything, mock.Anything).Return(scheduler.check, nil)

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	scheduler.msgBus = bus
	pm := secrets.NewProviderManager(&mockEventReceiver{})

	scheduler.scheduler = NewIntervalScheduler(ctx, s, scheduler.msgBus, scheduler.check, &cachev2.Resource{}, pm)

	assert.NoError(scheduler.msgBus.Start())

	switch executor {
	case "adhoc":
		scheduler.exec = NewAdhocRequestExecutor(ctx, s, &queue.Memory{}, scheduler.msgBus, &cachev2.Resource{}, pm)
	default:
		scheduler.exec = NewCheckExecutor(scheduler.msgBus, "default", s, &cachev2.Resource{}, pm)
	}

	return scheduler
}

func newCronScheduler(ctx context.Context, t *testing.T, executor string) *TestCronScheduler {
	t.Helper()

	assert := assert.New(t)

	scheduler := &TestCronScheduler{}
	scheduler.channel = make(chan interface{}, 2)

	request := corev2.FixtureCheckRequest("check1")
	asset := request.Assets[0]
	hook := request.Hooks[0]
	scheduler.check = request.Config
	scheduler.check.Interval = 1
	scheduler.check.Cron = "* * * * *"
	s := &mockstore.MockStore{}
	s.On("GetAssets", mock.Anything, &store.SelectionPredicate{}).Return([]*corev2.Asset{&asset}, nil)
	s.On("GetHookConfigs", mock.Anything, &store.SelectionPredicate{}).Return([]*corev2.HookConfig{&hook}, nil)
	s.On("GetCheckConfigByName", mock.Anything, mock.Anything).Return(scheduler.check, nil)

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	scheduler.msgBus = bus
	pm := secrets.NewProviderManager(&mockEventReceiver{})

	scheduler.scheduler = NewCronScheduler(ctx, s, scheduler.msgBus, scheduler.check, &cachev2.Resource{}, pm)

	assert.NoError(scheduler.msgBus.Start())

	switch executor {
	case "adhoc":
		scheduler.exec = NewAdhocRequestExecutor(ctx, s, &queue.Memory{}, scheduler.msgBus, &cachev2.Resource{}, pm)
	default:
		scheduler.exec = NewCheckExecutor(scheduler.msgBus, "default", s, &cachev2.Resource{}, pm)
	}

	return scheduler
}

func TestIntervalScheduling(t *testing.T) {
	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newIntervalScheduler(ctx, t, "check")

	// Set interval to smallest valid value
	check := scheduler.check
	check.Subscriptions = []string{"subscription1"}

	topic := messaging.SubscriptionTopic(check.Namespace, "subscription1")

	sub, err := scheduler.msgBus.Subscribe(topic, "scheduler", scheduler)
	if err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		_ = sub.Cancel()
		assert.NoError(scheduler.msgBus.Stop())
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		msg := <-scheduler.channel
		res, ok := msg.(*corev2.CheckRequest)
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
	scheduler := newIntervalScheduler(ctx, t, "check")

	// Set interval to smallest valid value
	check := scheduler.check
	check.Subscriptions = []string{"subscription1"}
	check.Subdues = []*corev2.TimeWindowRepeated{
		{
			Begin:  "2022-04-06T01:00:00-0400",
			End:    "2022-04-06T23:00:00-0400",
			Repeat: []string{corev2.RepeatPeriodDaily},
		}, {
			Begin:  "2022-04-06T22:00:00-0400",
			End:    "2022-04-07T02:00:00-0400",
			Repeat: []string{corev2.RepeatPeriodDaily},
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
		_ = subscription.Cancel()
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
	scheduler := newCronScheduler(ctx, t, "check")

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
		_ = subscription.Cancel()
		assert.NoError(scheduler.msgBus.Stop())
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		msg := <-scheduler.channel
		res, ok := msg.(*corev2.CheckRequest)
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
	scheduler := newCronScheduler(ctx, t, "check")

	// Set interval to smallest valid value
	check := scheduler.check
	check.Cron = "* * * * *"
	check.Subscriptions = []string{"subscription1"}
	check.Subdues = []*corev2.TimeWindowRepeated{
		{
			Begin:  "2022-04-06T01:00:00-0400",
			End:    "2022-04-06T23:00:00-0400",
			Repeat: []string{corev2.RepeatPeriodDaily},
		}, {
			Begin:  "2022-04-06T22:00:00-0400",
			End:    "2022-04-07T02:00:00-0400",
			Repeat: []string{corev2.RepeatPeriodDaily},
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
		_ = subscription.Cancel()
		assert.NoError(scheduler.msgBus.Stop())
	}()

	scheduler.scheduler.Start()
	mockTime.Set(mockTime.Now().Add(10 * time.Second))
	assert.NoError(scheduler.scheduler.Stop())

	// We should have no element in our channel
	assert.Equal(0, len(scheduler.channel))
}
