package schedulerd

import (
	"context"
	"sync"
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/secrets"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSchedulerd(t *testing.T) {
	// store stubs a check with interval scheduler
	intervalCheck := corev2.FixtureCheckConfig("interval")
	intervalCheck.Subscriptions = append(intervalCheck.Subscriptions, "disco")
	intervalCheck.Cron, intervalCheck.RoundRobin = "", false
	intervalCheck.Interval = 1

	// start schedulerd
	// observe bus messages
	stor := &mockstore.V2MockStore{}
	cs := &mockstore.ConfigStore{}
	cs.On(
		"List", mock.Anything, mock.MatchedBy(func(req storev2.ResourceRequest) bool { return req.Type == "CheckConfig" }), mock.Anything,
	).Return(
		mockstore.WrapList[*corev2.CheckConfig]([]*corev2.CheckConfig{intervalCheck}),
		nil,
	)
	cs.On(
		"List", mock.Anything, mock.MatchedBy(func(req storev2.ResourceRequest) bool { return req.Type == "Asset" }), mock.Anything,
	).Return(
		mockstore.WrapList[*corev2.Asset]([]*corev2.Asset{}),
		nil,
	)
	cs.On(
		"List", mock.Anything, mock.MatchedBy(func(req storev2.ResourceRequest) bool { return req.Type == "HookConfig" }), mock.Anything,
	).Return(
		mockstore.WrapList[*corev2.HookConfig]([]*corev2.HookConfig{}),
		nil,
	)
	es := &mockstore.EntityConfigStore{}
	es.On(
		"List", mock.Anything, mock.Anything, mock.Anything,
	).Return(
		[]*corev3.EntityConfig{},
		nil,
	)
	stor.On("GetConfigStore").Return(cs)
	stor.On("GetEntityConfigStore").Return(es)
	// subscribe to the wizard bus for check's sub
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start())

	discoC := make(chan interface{}, 10)
	discoS := testSubscriber{
		ch: discoC,
	}
	discoSub, err := bus.Subscribe(messaging.SubscriptionTopic("default", "disco"), "testing", discoS)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, discoSub.Cancel())
	}()

	sched, err := New(context.Background(), Config{
		Store:                  stor,
		Bus:                    bus,
		SecretsProviderManager: secrets.NewProviderManager(&mockEventReceiver{}),
	})
	require.NoError(t, err)
	require.NoError(t, sched.Start())
	mockTime.Start()
	defer mockTime.Stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		raw := <-discoC
		checkRequest, ok := raw.(*corev2.CheckRequest)
		assert.True(t, ok, "expected CheckRequest")
		assert.Equal(t, intervalCheck, checkRequest.Config)
		wg.Done()
	}()
	wg.Wait()
}
