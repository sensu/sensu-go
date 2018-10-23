// +build integration

package schedulerd

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/sensu/sensu-go/testing/mockring"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdhocExecutor(t *testing.T) {
	storeInst, err := testutil.NewStoreInstance()
	store := storeInst.GetStore()

	if err != nil {
		assert.FailNow(t, err.Error())
	}
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{
		RingGetter: &mockring.Getter{},
	})
	require.NoError(t, err)
	newAdhocExec := NewAdhocRequestExecutor(context.Background(), store, &queue.Memory{}, bus)
	defer newAdhocExec.Stop()
	assert.NoError(t, newAdhocExec.bus.Start())

	goodCheck := types.FixtureCheckConfig("goodCheck")
	goodCheck.Subscriptions = []string{"subscription1"}

	goodCheckRequest := &types.CheckRequest{}
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
	result, ok := msg.(*types.CheckRequest)
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
	scheduler := newScheduler(t, ctx)

	entity := types.FixtureEntity("entity1")
	check := scheduler.check
	check.Subscriptions = []string{"subscription1"}
	check.ProxyRequests = types.FixtureProxyRequests(true)

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
			res, ok := msg.(*types.CheckRequest)
			assert.True(ok)
			assert.Equal("check1", res.Config.Name)
			assert.Equal("entity1", res.Config.ProxyEntityID)
		}
	}()

	assert.NoError(scheduler.exec.publishProxyCheckRequests([]*types.Entity{entity}, check))
}

func TestPublishProxyCheckRequestsInterval(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scheduler := newScheduler(t, ctx)

	entity1 := types.FixtureEntity("entity1")
	entity2 := types.FixtureEntity("entity2")
	entity3 := types.FixtureEntity("entity3")
	entities := []*types.Entity{entity1, entity2, entity3}
	check := scheduler.check
	check.Subscriptions = []string{"subscription1"}
	check.ProxyRequests = types.FixtureProxyRequests(true)

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
				res, ok := msg.(*types.CheckRequest)
				assert.True(ok)
				assert.Equal("check1", res.Config.Name)
				assert.Equal(entityName, res.Config.ProxyEntityID)
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
	scheduler := newScheduler(t, ctx)

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
				res, ok := msg.(*types.CheckRequest)
				assert.True(ok)
				assert.Equal("check1", res.Config.Name)
				assert.Equal(entityName, res.Config.ProxyEntityID)
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
	scheduler := newScheduler(t, ctx)

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
	check.CheckHooks = []types.HookList{}
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
	scheduler := newScheduler(t, ctx)

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
	check.CheckHooks = []types.HookList{}
	request, err = scheduler.exec.buildRequest(check)
	require.NoError(t, err)
	assert.NotNil(request)
	assert.NotNil(request.Config)
	assert.Empty(request.Assets)
	assert.Empty(request.Hooks)

	assert.NoError(scheduler.msgBus.Stop())
}
