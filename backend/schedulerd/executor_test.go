// +build integration

package schedulerd

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store/cache"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdhocExecutor(t *testing.T) {
	store, err := testutil.NewStoreInstance()

	if err != nil {
		assert.FailNow(t, err.Error())
	}
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	pm := secrets.NewProviderManager()
	newAdhocExec := NewAdhocRequestExecutor(context.Background(), store, &queue.Memory{}, bus, &cache.Resource{}, pm)
	defer newAdhocExec.Stop()
	assert.NoError(t, newAdhocExec.bus.Start())

	goodCheck := corev2.FixtureCheckConfig("goodCheck")

	// set labels and annotations to nil to avoid value comparison issues
	goodCheck.Labels = nil
	goodCheck.Annotations = nil

	goodCheck.Subscriptions = []string{"subscription1"}

	goodCheckRequest := &corev2.CheckRequest{}
	goodCheckRequest.Config = goodCheck
	ch := make(chan interface{}, 1)
	tsub := testSubscriber{ch}

	topic := messaging.SubscriptionTopic(goodCheck.Namespace, "subscription1")
	sub, err := bus.Subscribe(topic, "testSubscriber", tsub)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer func() {
		close(ch)
		sub.Cancel()
	}()

	marshaledCheck, err := json.Marshal(goodCheck)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	if err = newAdhocExec.adhocQueue.Enqueue(context.Background(), string(marshaledCheck)); err != nil {
		assert.FailNow(t, err.Error())
	}

	msg := <-ch
	result, ok := msg.(*corev2.CheckRequest)
	assert.True(t, ok)
	assert.EqualValues(t, goodCheckRequest.Config, result.Config)
	assert.EqualValues(t, goodCheckRequest.Assets, result.Assets)
	assert.EqualValues(t, goodCheckRequest.Hooks, result.Hooks)
	assert.True(t, result.Issued > 0, "Issued > 0")
}

func TestPublishProxyCheckRequest(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newIntervalScheduler(ctx, t, "check")

	// Create two entities, so the first one fails token substitution, therefore
	// we can test that a check is scheduled for the second one
	entity1 := corev2.FixtureEntity("entity1")
	entity2 := corev2.FixtureEntity("entity2")
	entity2.Labels = map[string]string{"foo": "bar"}

	// Create a check that relies on token substitution
	check := scheduler.check
	check.Command = "{{ .labels.foo }}"
	check.Subscriptions = []string{"subscription1"}
	check.ProxyRequests = corev2.FixtureProxyRequests(true)

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf(
		"%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Namespace,
	)
	tsub := testSubscriber{
		ch: c1,
	}

	sub, err := scheduler.msgBus.Subscribe(topic, "testSubscriber", tsub)
	if err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		sub.Cancel()
		close(c1)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	go func() {
		select {
		case msg := <-c1:
			res, ok := msg.(*corev2.CheckRequest)
			assert.True(ok)
			assert.Equal("check1", res.Config.Name)
			assert.Equal("entity2", res.Config.ProxyEntityName)
		}
	}()

	assert.NoError(scheduler.exec.publishProxyCheckRequests([]*corev2.Entity{entity1, entity2}, check))
}

func TestPublishProxyCheckRequestsInterval(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newIntervalScheduler(ctx, t, "check")

	entity1 := corev2.FixtureEntity("entity1")
	entity2 := corev2.FixtureEntity("entity2")
	entity3 := corev2.FixtureEntity("entity3")
	entities := []*corev2.Entity{entity1, entity2, entity3}
	check := scheduler.check
	check.Subscriptions = []string{"subscription1"}
	check.ProxyRequests = corev2.FixtureProxyRequests(true)

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf(
		"%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Namespace,
	)

	tsub := testSubscriber{
		ch: c1,
	}

	sub, err := scheduler.msgBus.Subscribe(topic, "testSubscriber", tsub)
	if err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		sub.Cancel()
		close(c1)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	go func() {
		for i := 0; i < len(entities); i++ {
			entityName := fmt.Sprintf("entity%d", i+1)
			select {
			case msg := <-c1:
				res, ok := msg.(*corev2.CheckRequest)
				assert.True(ok)
				assert.Equal("check1", res.Config.Name)
				assert.Equal(entityName, res.Config.ProxyEntityName)
			}
		}
	}()

	assert.NoError(scheduler.exec.publishProxyCheckRequests(entities, check))
}

func TestPublishProxyCheckRequestsCron(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newCronScheduler(ctx, t, "check")

	entity1 := corev2.FixtureEntity("entity1")
	entity2 := corev2.FixtureEntity("entity2")
	entity3 := corev2.FixtureEntity("entity3")
	entities := []*corev2.Entity{entity1, entity2, entity3}
	check := scheduler.check
	check.Subscriptions = []string{"subscription1"}
	check.ProxyRequests = corev2.FixtureProxyRequests(true)
	check.Cron = "* * * * *"

	c1 := make(chan interface{}, 10)
	topic := fmt.Sprintf(
		"%s:%s:subscription1",
		messaging.TopicSubscriptions,
		check.Namespace,
	)

	tsub := testSubscriber{c1}

	sub, err := scheduler.msgBus.Subscribe(topic, "testSubscriber", tsub)
	if err != nil {
		assert.FailNow(err.Error())
	}
	defer func() {
		sub.Cancel()
		close(c1)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	go func() {
		for i := 0; i < len(entities); i++ {
			entityName := fmt.Sprintf("entity%d", i+1)
			select {
			case msg := <-c1:
				res, ok := msg.(*corev2.CheckRequest)
				assert.True(ok)
				assert.Equal("check1", res.Config.Name)
				assert.Equal(entityName, res.Config.ProxyEntityName)
			}
		}
	}()

	assert.NoError(scheduler.exec.publishProxyCheckRequests(entities, check))
}

func TestCheckBuildRequestInterval(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newIntervalScheduler(ctx, t, "check")

	check := scheduler.check
	request, err := scheduler.exec.buildRequest(check)
	require.NoError(t, err)
	assert.NotNil(request)
	assert.NotNil(request.Config)
	assert.NotNil(request.Assets)
	assert.NotEmpty(request.Assets)
	assert.Len(request.Assets, 1)
	assert.NotNil(request.Hooks)
	assert.NotEmpty(request.Hooks)
	assert.Len(request.Hooks, 1)

	check.RuntimeAssets = []string{}
	check.CheckHooks = []corev2.HookList{}
	request, err = scheduler.exec.buildRequest(check)
	require.NoError(t, err)
	assert.NotNil(request)
	assert.NotNil(request.Config)
	assert.Empty(request.Assets)
	assert.Empty(request.Hooks)
	assert.True(request.Issued > 0, "Issued > 0")

	assert.NoError(scheduler.msgBus.Stop())
}

func TestCheckBuildRequestCron(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newCronScheduler(ctx, t, "check")

	check := scheduler.check
	check.Cron = "* * * * *"

	request, err := scheduler.exec.buildRequest(check)
	require.NoError(t, err)
	assert.NotNil(request)
	assert.NotNil(request.Config)
	assert.NotNil(request.Assets)
	assert.NotEmpty(request.Assets)
	assert.Len(request.Assets, 1)
	assert.NotNil(request.Hooks)
	assert.NotEmpty(request.Hooks)
	assert.Len(request.Hooks, 1)

	check.RuntimeAssets = []string{}
	check.CheckHooks = []corev2.HookList{}
	request, err = scheduler.exec.buildRequest(check)
	require.NoError(t, err)
	assert.NotNil(request)
	assert.NotNil(request.Config)
	assert.Empty(request.Assets)
	assert.Empty(request.Hooks)

	assert.NoError(scheduler.msgBus.Stop())
}

func TestCheckBuildRequestAdhoc_GH2201(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newIntervalScheduler(ctx, t, "adhoc")

	check := scheduler.check
	check.Cron = "* * * * *"

	request, err := scheduler.exec.buildRequest(check)
	require.NoError(t, err)
	assert.NotNil(request)
	assert.NotNil(request.Config)
	assert.NotNil(request.Assets)
	assert.NotEmpty(request.Assets)
	assert.Len(request.Assets, 1)
	assert.NotNil(request.Hooks)
	assert.NotEmpty(request.Hooks)
	assert.Len(request.Hooks, 1)

	check.RuntimeAssets = []string{}
	check.CheckHooks = []corev2.HookList{}
	request, err = scheduler.exec.buildRequest(check)
	require.NoError(t, err)
	assert.NotNil(request)
	assert.NotNil(request.Config)
	assert.Empty(request.Assets)
	assert.Empty(request.Hooks)

	assert.NoError(scheduler.msgBus.Stop())
}
