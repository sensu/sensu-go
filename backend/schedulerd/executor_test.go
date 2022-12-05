package schedulerd

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
)

func TestPublishProxyCheckRequest(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newIntervalScheduler(ctx, t, "check")

	// Create two entities, so the first one fails token substitution, therefore
	// we can test that a check is scheduled for the second one
	entity1 := corev3.FixtureEntityConfig("entity1")
	entity2 := corev3.FixtureEntityConfig("entity2")
	entity2.Metadata.Labels = map[string]string{"foo": "bar"}

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
		assert.NoError(sub.Cancel())
		close(c1)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		msg := <-c1
		res, ok := msg.(*corev2.CheckRequest)
		assert.True(ok)
		assert.Equal("check1", res.Config.Name)
		assert.Equal("entity2", res.Config.ProxyEntityName)

	}()

	assert.NoError(scheduler.exec.publishProxyCheckRequests([]*corev3.EntityConfig{entity1, entity2}, check))

	wg.Wait()
}

func TestPublishProxyCheckRequestsInterval(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newIntervalScheduler(ctx, t, "check")

	entity1 := corev3.FixtureEntityConfig("entity1")
	entity2 := corev3.FixtureEntityConfig("entity2")
	entity3 := corev3.FixtureEntityConfig("entity3")
	entities := []*corev3.EntityConfig{entity1, entity2, entity3}
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
		assert.NoError(sub.Cancel())
		close(c1)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	go func() {
		for i := 0; i < len(entities); i++ {
			entityName := fmt.Sprintf("entity%d", i+1)
			msg := <-c1
			res, ok := msg.(*corev2.CheckRequest)
			assert.True(ok)
			assert.Equal("check1", res.Config.Name)
			assert.Equal(entityName, res.Config.ProxyEntityName)
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

	entity1 := corev3.FixtureEntityConfig("entity1")
	entity2 := corev3.FixtureEntityConfig("entity2")
	entity3 := corev3.FixtureEntityConfig("entity3")
	entities := []*corev3.EntityConfig{entity1, entity2, entity3}
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
		assert.NoError(sub.Cancel())
		close(c1)
		assert.NoError(scheduler.msgBus.Stop())
	}()

	go func() {
		for i := 0; i < len(entities); i++ {
			entityName := fmt.Sprintf("entity%d", i+1)
			msg := <-c1
			res, ok := msg.(*corev2.CheckRequest)
			assert.True(ok)
			assert.Equal("check1", res.Config.Name)
			assert.Equal(entityName, res.Config.ProxyEntityName)
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
