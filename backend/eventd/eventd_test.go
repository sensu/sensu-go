package eventd

import (
	"errors"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/monitor"
	"github.com/sensu/sensu-go/testing/mockmonitor"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEventHandling(t *testing.T) {
	bus := &messaging.WizardBus{}
	require.NoError(t, bus.Start())

	mockStore := &mockstore.MockStore{}
	e := &Eventd{
		Store:        mockStore,
		MessageBus:   bus,
		HandlerCount: 5,
	}

	err := e.Start()
	assert.NoError(t, err)

	require.NoError(t, bus.Publish(messaging.TopicEventRaw, nil))

	badEvent := &types.Event{}
	badEvent.Check = &types.Check{}
	badEvent.Entity = &types.Entity{}
	badEvent.Timestamp = time.Now().Unix()

	require.NoError(t, bus.Publish(messaging.TopicEventRaw, badEvent))

	event := types.FixtureEvent("entity", "check")

	var nilEvent *types.Event
	// no previous event.
	mockStore.On(
		"GetEventByEntityCheck",
		mock.Anything,
		"entity",
		"check",
	).Return(nilEvent, nil)
	mockStore.On("UpdateEvent", mock.AnythingOfType("*types.Event")).Return(nil)

	// No silenced entries
	mockStore.On(
		"GetSilencedEntriesBySubscription",
		mock.Anything,
	).Return([]*types.Silenced{}, nil)
	mockStore.On(
		"GetSilencedEntriesByCheckName",
		mock.Anything,
	).Return([]*types.Silenced{}, nil)

	require.NoError(t, bus.Publish(messaging.TopicEventRaw, event))

	err = e.Stop()
	assert.NoError(t, err)

	mockStore.AssertCalled(t, "UpdateEvent", mock.AnythingOfType("*types.Event"))

	// Make sure the event has been marked with the proper state
	assert.Equal(t, types.EventPassingState, event.Check.State)
	assert.Equal(t, event.Timestamp, event.Check.LastOK)
}

func TestEventMonitor(t *testing.T) {
	bus := &messaging.WizardBus{}
	require.NoError(t, bus.Start())

	mockStore := &mockstore.MockStore{}
	e := &Eventd{
		Store:        mockStore,
		MessageBus:   bus,
		HandlerCount: 5,
	}

	mon := &mockmonitor.MockMonitor{}
	mon.On("HandleUpdate", mock.Anything).Return(errors.New("error handling update"))
	e.MonitorFactory = func(*types.Entity, *types.Event, time.Duration, monitor.UpdateHandler, monitor.FailureHandler) monitor.Interface {
		return mon
	}

	err := e.Start()
	assert.NoError(t, err)

	require.NoError(t, bus.Publish(messaging.TopicEventRaw, nil))

	event := types.FixtureEvent("entity", "check")
	event.Check.Ttl = 90

	var nilEvent *types.Event
	// no previous event.
	mockStore.On(
		"GetEventByEntityCheck",
		mock.Anything,
		"entity",
		"check",
	).Return(nilEvent, nil)
	mockStore.On("UpdateEvent", mock.AnythingOfType("*types.Event")).Return(nil)

	// No silenced entries
	mockStore.On(
		"GetSilencedEntriesBySubscription",
		mock.Anything,
	).Return([]*types.Silenced{}, nil)
	mockStore.On(
		"GetSilencedEntriesByCheckName",
		mock.Anything,
	).Return([]*types.Silenced{}, nil)

	require.NoError(t, bus.Publish(messaging.TopicEventRaw, event))

	err = e.Stop()
	assert.NoError(t, err)

	mon.AssertCalled(t, "HandleUpdate", event)
	// Make sure the event has been marked with the proper state
	assert.Equal(t, types.EventPassingState, event.Check.State)
}
