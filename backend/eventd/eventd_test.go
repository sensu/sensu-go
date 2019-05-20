package eventd

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fakeSwitchSet struct{}

func (f *fakeSwitchSet) Alive(context.Context, string, int64) error {
	return nil
}

func (*fakeSwitchSet) Dead(context.Context, string, int64) error {
	return nil
}

func (*fakeSwitchSet) Bury(context.Context, string) error {
	return nil
}

func newFakeFactory(f liveness.Interface) liveness.Factory {
	return func(name string, dead, alive liveness.EventFunc, logger logrus.FieldLogger) liveness.Interface {
		return f
	}
}

func TestEventHandling(t *testing.T) {
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start())

	mockStore := &mockstore.MockStore{}
	e, err := New(Config{
		Store:           mockStore,
		Bus:             bus,
		LivenessFactory: newFakeFactory(&fakeSwitchSet{}),
	})
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
	mockStore.On("UpdateEvent", mock.Anything).Return(nil)

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

	mockStore.AssertCalled(t, "UpdateEvent", mock.Anything)

	assert.Equal(t, int64(1), event.Check.Occurrences)

	// Make sure the event has been marked with the proper state
	assert.Equal(t, types.EventPassingState, event.Check.State)
	assert.Equal(t, event.Timestamp, event.Check.LastOK)
}

func TestEventMonitor(t *testing.T) {
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start())

	mockStore := &mockstore.MockStore{}
	e, err := New(Config{Store: mockStore, Bus: bus})
	require.NoError(t, err)
	e.handlerCount = 5

	e.livenessFactory = newFakeFactory(&fakeSwitchSet{})

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
	mockStore.On("UpdateEvent", mock.Anything).Return(nil)

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

	// Make sure the event has been marked with the proper state
	assert.Equal(t, types.EventPassingState, event.Check.State)
}

func TestCheckOccurrences(t *testing.T) {
	OK := uint32(0)
	WARN := uint32(1)
	CRIT := uint32(2)

	// 1. Event with OK Check status - Occurrences: 1, OccurrencesWatermark: 1
	// 2. Event with OK Check status - Occurrences: 2, OccurrencesWatermark: 2
	// 3. Event with WARN Check status - Occurrences: 1, OccurrencesWatermark: 1
	// 4. Event with WARN Check status - Occurrences: 2, OccurrencesWatermark: 2
	// 5. Event with WARN Check status - Occurrences: 3, OccurrencesWatermark: 3
	// 6. Event with CRIT Check status - Occurrences: 1, OccurrencesWatermark: 3
	// 7. Event with CRIT Check status - Occurrences: 2, OccurrencesWatermark: 3
	// 8. Event with CRIT Check status - Occurrences: 3, OccurrencesWatermark: 3
	// 9. Event with CRIT Check status - Occurrences: 4, OccurrencesWatermark: 4
	// 10. Event with OK Check status - Occurrences: 1, OccurrencesWatermark: 4
	// 11. Event with CRIT Check status - Occurrences: 1, OccurrencesWatermark: 1
	testCases := []struct {
		name                         string
		status                       uint32
		history                      corev2.CheckHistory
		expectedOccurrences          int64
		expectedOccurrencesWatermark int64
	}{
		{
			name:                         "OK",
			status:                       OK,
			history:                      corev2.CheckHistory{Status: OK},
			expectedOccurrences:          1,
			expectedOccurrencesWatermark: 1,
		},
		{
			name:                         "OK -> OK",
			status:                       OK,
			history:                      corev2.CheckHistory{Status: OK},
			expectedOccurrences:          2,
			expectedOccurrencesWatermark: 2,
		},
		{
			name:                         "OK -> WARN",
			status:                       WARN,
			history:                      corev2.CheckHistory{Status: WARN},
			expectedOccurrences:          1,
			expectedOccurrencesWatermark: 1,
		},
		{
			name:                         "WARN -> WARN",
			status:                       WARN,
			history:                      corev2.CheckHistory{Status: WARN},
			expectedOccurrences:          2,
			expectedOccurrencesWatermark: 2,
		},
		{
			name:                         "WARN -> WARN",
			status:                       WARN,
			history:                      corev2.CheckHistory{Status: WARN},
			expectedOccurrences:          3,
			expectedOccurrencesWatermark: 3,
		},
		{
			name:                         "WARN -> CRIT",
			status:                       CRIT,
			history:                      corev2.CheckHistory{Status: CRIT},
			expectedOccurrences:          1,
			expectedOccurrencesWatermark: 3,
		},
		{
			name:                         "CRIT -> CRIT",
			status:                       CRIT,
			history:                      corev2.CheckHistory{Status: CRIT},
			expectedOccurrences:          2,
			expectedOccurrencesWatermark: 3,
		},
		{
			name:                         "CRIT -> CRIT",
			status:                       CRIT,
			history:                      corev2.CheckHistory{Status: CRIT},
			expectedOccurrences:          3,
			expectedOccurrencesWatermark: 3,
		},
		{
			name:                         "CRIT -> CRIT",
			status:                       CRIT,
			history:                      corev2.CheckHistory{Status: CRIT},
			expectedOccurrences:          4,
			expectedOccurrencesWatermark: 4,
		},
		{
			name:                         "CRIT -> OK",
			status:                       OK,
			history:                      corev2.CheckHistory{Status: OK},
			expectedOccurrences:          1,
			expectedOccurrencesWatermark: 4,
		},
		{
			name:                         "OK -> CRIT",
			status:                       CRIT,
			history:                      corev2.CheckHistory{Status: CRIT},
			expectedOccurrences:          1,
			expectedOccurrencesWatermark: 1,
		},
	}

	event := corev2.FixtureEvent("entity1", "check1")
	event.Check.Occurrences = 1
	event.Check.OccurrencesWatermark = 1
	event.Check.History = []corev2.CheckHistory{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event.Check.Status = tc.status
			event.Check.History = append(event.Check.History, tc.history)
			updateOccurrences(event)
			assert.Equal(t, tc.expectedOccurrences, event.Check.Occurrences)
			assert.Equal(t, tc.expectedOccurrencesWatermark, event.Check.OccurrencesWatermark)
		})
	}
}

type mockSwitchSet struct {
	mock.Mock
}

func (m *mockSwitchSet) Alive(ctx context.Context, id string, ttl int64) error {
	args := m.Called(ctx, id, ttl)
	return args.Error(0)
}

func (m *mockSwitchSet) Dead(ctx context.Context, id string, ttl int64) error {
	args := m.Called(ctx, id, ttl)
	return args.Error(0)
}

func (m *mockSwitchSet) Bury(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCheckTTL(t *testing.T) {
	ttlCheck := corev2.FixtureEvent("entity", "check")
	ttlCheck.Check.Ttl = 120
	nonTTLCheck := corev2.FixtureEvent("entity", "check")

	type switchesFunc func(*mockSwitchSet)

	tests := []struct {
		name             string
		msg              interface{}
		previousEvent    *corev2.Event
		previousEventErr error
		switchesFunc     switchesFunc
		wantErr          bool
	}{
		{
			name:    "a check without TTL shouldn't bury its switchset",
			msg:     nonTTLCheck,
			wantErr: false,
		},
		{
			name: "a check with TTL should update its switchset with alive",
			msg:  ttlCheck,
			switchesFunc: func(s *mockSwitchSet) {
				s.On("Alive", mock.Anything, "default/check/entity", int64(120)).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "switchset alive err should be returned",
			msg:  ttlCheck,
			switchesFunc: func(s *mockSwitchSet) {
				s.On("Alive", mock.Anything, "default/check/entity", int64(120)).
					Return(errors.New("err"))
			},
			wantErr: true,
		},
		{
			name:          "a check that just got its TTL disabled should bury its switchet",
			msg:           nonTTLCheck,
			previousEvent: ttlCheck,
			switchesFunc: func(s *mockSwitchSet) {
				s.On("Bury", mock.Anything, "default/check/entity").Return(nil)
			},
			wantErr: false,
		},
		{
			name:          "an error while burying the switchset should not be returned",
			msg:           nonTTLCheck,
			previousEvent: ttlCheck,
			switchesFunc: func(s *mockSwitchSet) {
				s.On("Bury", mock.Anything, "default/check/entity").Return(errors.New("err"))
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockstore.MockStore{}
			switches := &mockSwitchSet{}
			if tt.switchesFunc != nil {
				tt.switchesFunc(switches)
			}

			e := &Eventd{
				store:           store,
				livenessFactory: newFakeFactory(switches),
				handlerCount:    1,
				wg:              &sync.WaitGroup{},
			}
			var err error
			e.bus, err = messaging.NewWizardBus(messaging.WizardBusConfig{})
			require.NoError(t, err)
			require.NoError(t, e.bus.Start())

			store.On("GetEventByEntityCheck", mock.Anything, "entity", "check").
				Return(tt.previousEvent, tt.previousEventErr)
			store.On("GetSilencedEntriesBySubscription", mock.Anything, mock.Anything).
				Return([]*corev2.Silenced{}, nil)
			store.On("GetSilencedEntriesByCheckName", mock.Anything, mock.Anything).
				Return([]*corev2.Silenced{}, nil)
			store.On("UpdateEvent", mock.Anything, mock.Anything).Return(nil)

			if err := e.handleMessage(tt.msg); (err != nil) != tt.wantErr {
				t.Errorf("Eventd.handleMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
