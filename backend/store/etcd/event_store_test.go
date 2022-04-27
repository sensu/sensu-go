package etcd

import (
	"context"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
)

func newEventFixture(entity, check string) *corev2.Event {
	event := corev2.FixtureEvent(entity, check)
	event.Check.State = ""
	event.Check.LastOK = 0
	event.Check.Status = 1
	event.Check.Issued = 1610056762
	event.Check.Executed = 1610056763
	event.Check.History = []corev2.CheckHistory{}
	return event
}

func Test_updateEventHistory(t *testing.T) {
	type eventFn func() *corev2.Event

	tests := []struct {
		name        string
		eventFn     eventFn
		prevEventFn eventFn
		wantEventFn eventFn
		wantErr     bool
		errMatch    string
	}{
		{
			name: "previous event exists with no check",
			eventFn: func() *corev2.Event {
				return newEventFixture("foo", "bar")
			},
			prevEventFn: func() *corev2.Event {
				event := corev2.FixtureEvent("foo", "bar")
				event.Check = nil
				return event
			},
			wantErr:  true,
			errMatch: "invalid previous event",
		},
		{
			name: "previous event exists",
			eventFn: func() *corev2.Event {
				return newEventFixture("foo", "bar")
			},
			prevEventFn: func() *corev2.Event {
				event := newEventFixture("foo", "bar")
				event.Check.State = corev2.EventFailingState
				event.Check.Issued = 1610056752
				event.Check.Executed = 1610056753
				event.Check.LastOK = 1610056743
				event.Check.History = []corev2.CheckHistory{
					{Status: 1, Executed: 1610056753},
				}
				return event
			},
			wantEventFn: func() *corev2.Event {
				event := newEventFixture("foo", "bar")
				event.Check.State = corev2.EventFailingState
				event.Check.LastOK = 1610056743
				event.Check.History = []corev2.CheckHistory{
					{Status: 1, Executed: 1610056753},
					{Status: 1, Executed: 1610056763},
				}
				return event
			},
			wantErr: false,
		},
		{
			name: "no previous event exists, check status 1",
			eventFn: func() *corev2.Event {
				return newEventFixture("foo", "bar")
			},
			prevEventFn: func() *corev2.Event {
				return nil
			},
			wantEventFn: func() *corev2.Event {
				event := newEventFixture("foo", "bar")
				event.Check.State = corev2.EventFailingState
				event.Check.LastOK = 0
				event.Check.History = []corev2.CheckHistory{
					{Status: 1, Executed: 1610056763},
				}
				return event
			},
			wantErr: false,
		},
		{
			name: "no previous event exists, check status 0",
			eventFn: func() *corev2.Event {
				event := newEventFixture("foo", "bar")
				event.Check.Status = 0
				return event
			},
			prevEventFn: func() *corev2.Event {
				return nil
			},
			wantEventFn: func() *corev2.Event {
				event := newEventFixture("foo", "bar")
				event.Check.State = corev2.EventPassingState
				event.Check.LastOK = 1610056763
				event.Check.History = []corev2.CheckHistory{
					{Status: 0, Executed: 1610056763},
				}
				return event
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := tt.eventFn()
			prevEvent := tt.prevEventFn()
			err := updateEventHistory(event, prevEvent)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateEventHistory() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.errMatch != "" {
				if err != nil {
					assert.Contains(t, err.Error(), tt.errMatch)
				} else {
					assert.Contains(t, "", tt.errMatch)
				}
			}
			if tt.wantEventFn != nil {
				want := tt.wantEventFn()
				if !reflect.DeepEqual(want.Check.History, event.Check.History) {
					t.Errorf("updateEventHistory() want history %v, got %v", want.Check.History, event.Check.History)
				}
				if want.Check.State != event.Check.State {
					t.Errorf("updateEventHistory() want state %v, got %v", want.Check.State, event.Check.State)
				}
				if want.Check.LastOK != event.Check.LastOK {
					t.Errorf("updateEventHistory() want last_ok %v, got %v", want.Check.LastOK, event.Check.LastOK)
				}
			}
		})
	}
}

func TestEventStoreSupportsFilteringUnsupported(t *testing.T) {
	store := NewStore(nil)
	assert.Equal(t, false, store.EventStoreSupportsFiltering(context.Background()), "etcd event store not expected to support filtering")
}
