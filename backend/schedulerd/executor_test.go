// +build integration,race

package schedulerd

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store/etcd/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestAdhocExecutor(t *testing.T) {
	store, err := testutil.NewStoreInstance()

	if err != nil {
		assert.FailNow(t, err.Error())
	}
	bus := &messaging.WizardBus{}
	newAdhocExec := NewAdhocRequestExecutor(context.Background(), store, bus)
	defer newAdhocExec.Stop()
	assert.NoError(t, newAdhocExec.bus.Start())

	goodCheck := types.FixtureCheckConfig("goodCheck")
	goodCheck.Subscriptions = []string{"subscription1"}

	goodCheckRequest := &types.CheckRequest{}
	goodCheckRequest.Config = goodCheck
	ch := make(chan interface{}, 1)
	topic := messaging.SubscriptionTopic(goodCheck.Organization, goodCheck.Environment, "subscription1")
	assert.NoError(t, bus.Subscribe(topic, "channel", ch))

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
	assert.EqualValues(t, goodCheckRequest, result)
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

	assert.NoError(scheduler.exec.publishProxyCheckRequests([]*types.Entity{entity}, check))
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

	assert.NoError(scheduler.exec.publishProxyCheckRequests(entities, check))
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

	assert.NoError(scheduler.exec.publishProxyCheckRequests(entities, check))
}

func TestCheckBuildRequestInterval(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	scheduler := newScheduler(t)

	check := scheduler.check
	request := scheduler.exec.buildRequest(check)
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
	request = scheduler.exec.buildRequest(check)
	assert.NotNil(request)
	assert.NotNil(request.Config)
	assert.Empty(request.Assets)
	assert.Empty(request.Hooks)

	assert.NoError(scheduler.msgBus.Stop())
}

func TestCheckBuildRequestCron(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Start a scheduler
	scheduler := newScheduler(t)

	check := scheduler.check
	check.Cron = "* * * * *"

	request := scheduler.exec.buildRequest(check)
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
	request = scheduler.exec.buildRequest(check)
	assert.NotNil(request)
	assert.NotNil(request.Config)
	assert.Empty(request.Assets)
	assert.Empty(request.Hooks)

	assert.NoError(scheduler.msgBus.Stop())
}
