package pipeline

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/rpc"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPipelineFilter(t *testing.T) {
	p := &Pipeline{
		eventClient: new(mockEventClient),
	}
	store := &mockstore.MockStore{}
	p.store = store

	// Mock the store responses
	allowFilterBar := &types.EventFilter{
		ObjectMeta: types.ObjectMeta{
			Name: "allowFilterBar",
		},
		Action:      types.EventFilterActionAllow,
		Expressions: []string{`event.check.output == "bar"`},
	}
	allowFilterFoo := &types.EventFilter{
		ObjectMeta: types.ObjectMeta{
			Name: "allowFilterFoo",
		},
		Action:      types.EventFilterActionAllow,
		Expressions: []string{`event.check.output == "foo"`},
	}
	denyFilterBar := &types.EventFilter{
		ObjectMeta: types.ObjectMeta{
			Name: "denyFilterBar",
		},
		Action:      types.EventFilterActionDeny,
		Expressions: []string{`event.check.output == "bar"`},
	}
	denyFilterFoo := &types.EventFilter{
		ObjectMeta: types.ObjectMeta{
			Name: "denyFilterFoo",
		},
		Action:      types.EventFilterActionDeny,
		Expressions: []string{`event.check.output == "foo"`},
	}
	store.On("GetEventFilterByName", mock.Anything, "allowFilterBar").Return(allowFilterBar, nil)
	store.On("GetEventFilterByName", mock.Anything, "allowFilterFoo").Return(allowFilterFoo, nil)
	store.On("GetEventFilterByName", mock.Anything, "denyFilterBar").Return(denyFilterBar, nil)
	store.On("GetEventFilterByName", mock.Anything, "denyFilterFoo").Return(denyFilterFoo, nil)
	store.On("GetEventFilterByName", mock.Anything, "extension_filter").Return((*types.EventFilter)(nil), nil)
	store.On("GetExtension", mock.Anything, "extension_filter").Return(&types.Extension{URL: "http://127.0.0.1"}, nil)

	p.extensionExecutor = func(ext *types.Extension) (rpc.ExtensionExecutor, error) {
		m := &mockExec{}
		m.On("FilterEvent", mock.Anything).Return(true, nil)
		return m, nil
	}

	testCases := []struct {
		name           string
		status         uint32
		history        []types.CheckHistory
		metrics        *types.Metrics
		silenced       []string
		filters        []string
		expectedFilter string
	}{
		{
			name:           "Not Incident",
			status:         0,
			metrics:        nil,
			silenced:       []string{},
			filters:        []string{"is_incident"},
			expectedFilter: "is_incident",
		},
		{
			name:           "Incident",
			status:         1,
			metrics:        nil,
			silenced:       []string{},
			filters:        []string{"is_incident"},
			expectedFilter: "",
		},
		{
			name:           "Metrics OK Status",
			status:         0,
			metrics:        &types.Metrics{},
			silenced:       []string{},
			filters:        []string{"has_metrics"},
			expectedFilter: "",
		},
		{
			name:           "Metrics Warning Status",
			status:         1,
			metrics:        &types.Metrics{},
			silenced:       []string{},
			filters:        []string{"has_metrics"},
			expectedFilter: "",
		},
		{
			name:           "Allow Filter With No Match",
			status:         1,
			metrics:        nil,
			silenced:       []string{},
			filters:        []string{"allowFilterBar"},
			expectedFilter: "allowFilterBar",
		},
		{
			name:           "Allow Filter With Match",
			status:         1,
			metrics:        nil,
			silenced:       []string{},
			filters:        []string{"allowFilterFoo"},
			expectedFilter: "",
		},
		{
			name:           "Deny Filter With No Match",
			status:         1,
			metrics:        nil,
			silenced:       []string{},
			filters:        []string{"denyFilterBar"},
			expectedFilter: "",
		},
		{
			name:           "Deny Filter With Match",
			status:         1,
			metrics:        nil,
			silenced:       []string{},
			filters:        []string{"denyFilterFoo"},
			expectedFilter: "denyFilterFoo",
		},
		{
			name:           "Deny Filters With One Match",
			status:         1,
			metrics:        nil,
			silenced:       []string{},
			filters:        []string{"denyFilterBar", "denyFilterFoo"},
			expectedFilter: "denyFilterFoo",
		},
		{
			name:           "Silenced With Metrics",
			status:         1,
			metrics:        &types.Metrics{},
			silenced:       []string{"entity1"},
			filters:        []string{"has_metrics"},
			expectedFilter: "",
		},
		{
			name:           "Silenced Without Metrics",
			status:         1,
			metrics:        nil,
			silenced:       []string{"entity1"},
			filters:        []string{"is_incident", "not_silenced"},
			expectedFilter: "not_silenced",
		},
		{
			name:   "Not Transitioned From Incident To Healthy",
			status: 0,
			history: []types.CheckHistory{
				types.CheckHistory{Status: 0},
				types.CheckHistory{Status: 0},
			},
			metrics:        nil,
			silenced:       []string{},
			filters:        []string{"is_incident"},
			expectedFilter: "is_incident",
		},
		{
			name:   "Transitioned From Incident To Healthy",
			status: 0,
			history: []types.CheckHistory{
				types.CheckHistory{Status: 1},
				types.CheckHistory{Status: 0},
			},
			metrics:        nil,
			silenced:       []string{},
			filters:        []string{"is_incident"},
			expectedFilter: "",
		},
		{
			name:           "Extension filter",
			filters:        []string{"extension_filter"},
			expectedFilter: "extension_filter",
		},
	}

	for _, tc := range testCases {
		handler := &types.Handler{
			Type:    "pipe",
			Command: "cat",
			Filters: tc.filters,
		}

		t.Run(tc.name, func(t *testing.T) {
			event := &types.Event{
				Check: &types.Check{
					Status:   tc.status,
					History:  tc.history,
					Output:   "foo",
					Silenced: tc.silenced,
				},
				Entity: &types.Entity{
					ObjectMeta: types.ObjectMeta{
						Namespace: "default",
					},
				},
				Metrics: tc.metrics,
			}

			f, _ := p.filterEvent(handler, event)
			assert.Equal(t, tc.expectedFilter, f)
		})
	}
}

type mockEventClient struct {
	mock.Mock
}

func (m *mockEventClient) FetchEvent(ctx context.Context, namespace, name string) (*corev2.Event, error) {
	args := m.Called(ctx, namespace, name)
	return args.Get(0).(*corev2.Event), args.Error(1)
}

func (m *mockEventClient) ListEvents(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Event, error) {
	args := m.Called(ctx, pred)
	return args.Get(0).([]*corev2.Event), args.Error(1)
}

func TestPipelineWhenFilter(t *testing.T) {
	p := &Pipeline{
		eventClient: new(mockEventClient),
	}
	store := &mockstore.MockStore{}
	p.store = store

	event := &types.Event{
		Check: &types.Check{
			Status: 2,
			Output: "bar",
		},
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Namespace: "default",
			},
		},
	}

	testCases := []struct {
		name           string
		filterName     string
		action         string
		begin          time.Duration
		end            time.Duration
		expectedFilter string
	}{
		{
			name:           "in time window action allow",
			filterName:     "in_time_window_allow",
			action:         types.EventFilterActionAllow,
			begin:          -time.Minute * time.Duration(1),
			end:            time.Minute * time.Duration(1),
			expectedFilter: "",
		},
		{
			name:           "outside time window action allow",
			filterName:     "outside_time_window_allow",
			action:         types.EventFilterActionAllow,
			begin:          time.Minute * time.Duration(10),
			end:            time.Minute * time.Duration(20),
			expectedFilter: "outside_time_window_allow",
		},
		{
			name:           "in time window action deny",
			filterName:     "in_time_window_deny",
			action:         types.EventFilterActionDeny,
			begin:          -time.Minute * time.Duration(1),
			end:            time.Minute * time.Duration(1),
			expectedFilter: "in_time_window_deny",
		},
		{
			name:           "outside time window action deny",
			filterName:     "outside_time_window_deny",
			action:         types.EventFilterActionDeny,
			begin:          time.Minute * time.Duration(10),
			end:            time.Minute * time.Duration(20),
			expectedFilter: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now().UTC()

			filter := &types.EventFilter{
				ObjectMeta: types.ObjectMeta{
					Name: tc.name,
				},
				Action:      tc.action,
				Expressions: []string{`event.check.output == "bar"`},
			}

			filter.When = &types.TimeWindowWhen{
				Days: types.TimeWindowDays{
					All: []*types.TimeWindowTimeRange{{
						Begin: now.Add(tc.begin).Format("03:04PM"),
						End:   now.Add(tc.end).Format("03:04PM"),
					}},
				},
			}

			store.On("GetEventFilterByName", mock.Anything, tc.filterName).Return(filter, nil)

			handler := &types.Handler{
				Type:    "pipe",
				Command: "cat",
				Filters: []string{tc.filterName},
			}

			f, _ := p.filterEvent(handler, event)
			assert.Equal(t, tc.expectedFilter, f)
		})
	}
}

func TestPipelineEventClient(t *testing.T) {
	events := new(mockEventClient)
	events.On("FetchEvent", mock.Anything, "foo", "bar").Return(corev2.FixtureEvent("foo", "bar"), nil)
	events.On("FetchEvent", mock.Anything, "foo", "baz").Return((*corev2.Event)(nil), &store.ErrNotFound{Key: "baz"})
	p := &Pipeline{
		eventClient: events,
	}
	store := &mockstore.MockStore{}
	p.store = store

	const getEvent = `(function () {
		var event = sensu.FetchEvent("%s", "%s");
		return event.Check.Status > 0;
	})();`

	getEventFromStore := &types.EventFilter{
		ObjectMeta: types.ObjectMeta{
			Name: "filter1",
		},
		Action:      types.EventFilterActionAllow,
		Expressions: []string{fmt.Sprintf(getEvent, "foo", "bar")},
	}
	getEventFromStoreMissing := &types.EventFilter{
		ObjectMeta: types.ObjectMeta{
			Name: "filter2",
		},
		Action:      types.EventFilterActionAllow,
		Expressions: []string{fmt.Sprintf(getEvent, "foo", "baz")},
	}
	store.On("GetEventFilterByName", mock.Anything, "filter1").Return(getEventFromStore, nil)
	store.On("GetEventFilterByName", mock.Anything, "filter2").Return(getEventFromStoreMissing, nil)
	p.extensionExecutor = func(ext *types.Extension) (rpc.ExtensionExecutor, error) {
		m := &mockExec{}
		m.On("FilterEvent", mock.Anything).Return(true, nil)
		return m, nil
	}

	handler := &types.Handler{
		Type:    "pipe",
		Command: "cat",
		Filters: []string{"filter1"},
	}
	handlerErr := &corev2.Handler{
		Type:    "pipe",
		Command: "cat",
		Filters: []string{"filter2"},
	}

	event := &types.Event{
		Check: &types.Check{
			Output: "foo",
		},
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Namespace: "default",
			},
		},
	}

	fname, err := p.filterEvent(handler, event)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := fname, "filter1"; got != want {
		t.Errorf("bad filter: got %s, want %s", got, want)
	}

	fname, err = p.filterEvent(handlerErr, event)
	if err != nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := fname, ""; got != want {
		t.Error("expected no filter match for broken filter")
	}
}
