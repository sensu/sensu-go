package eventd

import (
	"errors"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/monitor"
	"github.com/sensu/sensu-go/testing/mockmonitor"
	"github.com/sensu/sensu-go/testing/mockring"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEventHandling(t *testing.T) {
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{
		RingGetter: &mockring.Getter{},
	})
	require.NoError(t, err)
	require.NoError(t, bus.Start())

	mockStore := &mockstore.MockStore{}
	e, err := New(Config{Store: mockStore, Bus: bus})
	require.NoError(t, err)
	e.handlerCount = 5

	require.NoError(t, e.Start())

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

	assert.Equal(t, int64(1), event.Check.Occurrences)

	// Make sure the event has been marked with the proper state
	assert.Equal(t, types.EventPassingState, event.Check.State)
	assert.Equal(t, event.Timestamp, event.Check.LastOK)
}

func TestEventMonitor(t *testing.T) {
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{
		RingGetter: &mockring.Getter{},
	})
	require.NoError(t, err)
	require.NoError(t, bus.Start())

	mockStore := &mockstore.MockStore{}
	e, err := New(Config{Store: mockStore, Bus: bus})
	require.NoError(t, err)
	e.handlerCount = 5

	mon := &mockmonitor.MockMonitor{}
	mon.On("HandleUpdate", mock.Anything).Return(errors.New("error handling update"))
	e.monitorFactory = func(*types.Entity, *types.Event, time.Duration, monitor.UpdateHandler, monitor.FailureHandler) monitor.Interface {
		return mon
	}

	require.NoError(t, e.Start())

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

func TestCheckOccurrences(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name                         string
		status                       uint32
		occurrences                  int64
		history                      []types.CheckHistory
		expectedOccurrences          int64
		expectedOccurrencesWatermark int64
	}{
		{
			name:        "No previous occurences, check OK",
			status:      0,
			occurrences: int64(0),
			history: []types.CheckHistory{
				{Status: 0, Executed: now.Unix() - 1},
			},
			expectedOccurrences:          1,
			expectedOccurrencesWatermark: 1,
		},
		{
			name:        "No previous occurences, check WARN",
			status:      1,
			occurrences: int64(0),
			history: []types.CheckHistory{
				{Status: 1, Executed: now.Unix() - 1},
			},
			expectedOccurrences:          1,
			expectedOccurrencesWatermark: 1,
		},
		{
			name:        "previous WARN occurences, check OK",
			status:      0,
			occurrences: int64(1),
			history: []types.CheckHistory{
				{Status: 1, Executed: now.Unix() - 2},
				{Status: 0, Executed: now.Unix() - 1},
			},
			expectedOccurrences:          1,
			expectedOccurrencesWatermark: 1,
		},
		{
			name:        "previous WARN occurences, check WARN",
			status:      1,
			occurrences: int64(1),
			history: []types.CheckHistory{
				{Status: 1, Executed: now.Unix() - 2},
				{Status: 1, Executed: now.Unix() - 1},
			},
			expectedOccurrences:          2,
			expectedOccurrencesWatermark: 2,
		},
		{
			name:        "previous CRIT occurences, check WARN",
			status:      1,
			occurrences: int64(1),
			history: []types.CheckHistory{
				{Status: 2, Executed: now.Unix() - 2},
				{Status: 1, Executed: now.Unix() - 1},
			},
			expectedOccurrences:          1,
			expectedOccurrencesWatermark: 1,
		},
		{
			name:        "GH-1624 occurrences property should not reset to 1 on resolve",
			status:      0,
			occurrences: 5,
			history: []types.CheckHistory{
				{Status: 2, Executed: now.Unix() - 6},
				{Status: 2, Executed: now.Unix() - 5},
				{Status: 2, Executed: now.Unix() - 4},
				{Status: 2, Executed: now.Unix() - 3},
				{Status: 2, Executed: now.Unix() - 2},
			},
			expectedOccurrences:          5,
			expectedOccurrencesWatermark: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := types.FixtureEvent("entity1", "check1")
			event.Check.Status = tc.status
			event.Check.Occurrences = tc.occurrences
			event.Check.History = tc.history
			updateOccurrences(event)

			assert.Equal(t, tc.expectedOccurrences, event.Check.Occurrences)
			assert.Equal(t, tc.expectedOccurrencesWatermark, event.Check.OccurrencesWatermark)
		})
	}
}
