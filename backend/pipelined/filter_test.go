// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"testing"
	"time"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPipelinedFilter(t *testing.T) {
	p := &Pipelined{}
	store := &mockstore.MockStore{}
	p.Store = store

	// Mock the store responses
	allowFilterBar := &types.EventFilter{
		Name:       "allowFilterBar",
		Action:     types.EventFilterActionAllow,
		Statements: []string{`event.Check.Output == "bar"`},
	}
	allowFilterFoo := &types.EventFilter{
		Name:       "allowFilterFoo",
		Action:     types.EventFilterActionAllow,
		Statements: []string{`event.Check.Output == "foo"`},
	}
	denyFilterBar := &types.EventFilter{
		Name:       "denyFilterBar",
		Action:     types.EventFilterActionDeny,
		Statements: []string{`event.Check.Output == "bar"`},
	}
	denyFilterFoo := &types.EventFilter{
		Name:       "denyFilterFoo",
		Action:     types.EventFilterActionDeny,
		Statements: []string{`event.Check.Output == "foo"`},
	}
	store.On("GetEventFilterByName", mock.Anything, "allowFilterBar").Return(allowFilterBar, nil)
	store.On("GetEventFilterByName", mock.Anything, "allowFilterFoo").Return(allowFilterFoo, nil)
	store.On("GetEventFilterByName", mock.Anything, "denyFilterBar").Return(denyFilterBar, nil)
	store.On("GetEventFilterByName", mock.Anything, "denyFilterFoo").Return(denyFilterFoo, nil)

	testCases := []struct {
		name     string
		status   int32
		history  []types.CheckHistory
		metrics  *types.Metrics
		silenced []string
		filters  []string
		expected bool
	}{
		{
			name:     "Not Incident",
			status:   0,
			metrics:  nil,
			silenced: []string{},
			filters:  []string{"is_incident"},
			expected: true,
		},
		{
			name:     "Incident",
			status:   1,
			metrics:  nil,
			silenced: []string{},
			filters:  []string{"is_incident"},
			expected: false,
		},
		{
			name:     "Metrics OK Status",
			status:   0,
			metrics:  &types.Metrics{},
			silenced: []string{},
			filters:  []string{"has_metrics"},
			expected: false,
		},
		{
			name:     "Metrics Warning Status",
			status:   1,
			metrics:  &types.Metrics{},
			silenced: []string{},
			filters:  []string{"has_metrics"},
			expected: false,
		},
		{
			name:     "Allow Filter With No Match",
			status:   1,
			metrics:  nil,
			silenced: []string{},
			filters:  []string{"allowFilterBar"},
			expected: true,
		},
		{
			name:     "Allow Filter With Match",
			status:   1,
			metrics:  nil,
			silenced: []string{},
			filters:  []string{"allowFilterFoo"},
			expected: false,
		},
		{
			name:     "Deny Filter With No Match",
			status:   1,
			metrics:  nil,
			silenced: []string{},
			filters:  []string{"denyFilterBar"},
			expected: false,
		},
		{
			name:     "Deny Filter With Match",
			status:   1,
			metrics:  nil,
			silenced: []string{},
			filters:  []string{"denyFilterFoo"},
			expected: true,
		},
		{
			name:     "Deny Filters With One Match",
			status:   1,
			metrics:  nil,
			silenced: []string{},
			filters:  []string{"denyFilterBar", "denyFilterFoo"},
			expected: true,
		},
		{
			name:     "Silenced With Metrics",
			status:   1,
			metrics:  &types.Metrics{},
			silenced: []string{"entity1"},
			filters:  []string{"has_metrics"},
			expected: false,
		},
		{
			name:     "Silenced Without Metrics",
			status:   1,
			metrics:  nil,
			silenced: []string{"entity1"},
			filters:  []string{"is_incident", "not_silenced"},
			expected: true,
		},
		{
			name:   "Not Transitioned From Incident To Healthy",
			status: 0,
			history: []types.CheckHistory{
				types.CheckHistory{Status: 0},
			},
			metrics:  nil,
			silenced: []string{},
			filters:  []string{"is_incident"},
			expected: true,
		},
		{
			name:   "Transitioned From Incident To Healthy",
			status: 0,
			history: []types.CheckHistory{
				types.CheckHistory{Status: 1},
			},
			metrics:  nil,
			silenced: []string{},
			filters:  []string{"is_incident"},
			expected: false,
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
					Status:  tc.status,
					History: tc.history,
					Output:  "foo",
				},
				Entity: &types.Entity{
					Environment:  "default",
					Organization: "default",
				},
				Metrics:  tc.metrics,
				Silenced: tc.silenced,
			}

			filtered := p.filterEvent(handler, event)
			assert.Equal(t, tc.expected, filtered)
		})
	}
}

func TestPipelinedWhenFilter(t *testing.T) {
	p := &Pipelined{}
	store := &mockstore.MockStore{}
	p.Store = store

	event := &types.Event{
		Check: &types.Check{
			Status: 2,
			Output: "bar",
		},
		Entity: &types.Entity{
			Environment:  "default",
			Organization: "default",
		},
	}

	testCases := []struct {
		name       string
		filterName string
		action     string
		begin      time.Duration
		end        time.Duration
		expected   bool
	}{
		{
			name:       "in time window action allow",
			filterName: "in_time_window_allow",
			action:     types.EventFilterActionAllow,
			begin:      -time.Minute * time.Duration(1),
			end:        time.Minute * time.Duration(1),
			expected:   false,
		},
		{
			name:       "outside time window action allow",
			filterName: "outside_time_window_allow",
			action:     types.EventFilterActionAllow,
			begin:      time.Minute * time.Duration(10),
			end:        time.Minute * time.Duration(20),
			expected:   true,
		},
		{
			name:       "in time window action deny",
			filterName: "in_time_window_deny",
			action:     types.EventFilterActionDeny,
			begin:      -time.Minute * time.Duration(1),
			end:        time.Minute * time.Duration(1),
			expected:   true,
		},
		{
			name:       "outside time window action deny",
			filterName: "outside_time_window_deny",
			action:     types.EventFilterActionDeny,
			begin:      time.Minute * time.Duration(10),
			end:        time.Minute * time.Duration(20),
			expected:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now().UTC()

			filter := &types.EventFilter{
				Name:       tc.name,
				Action:     tc.action,
				Statements: []string{`event.Check.Output == "bar"`},
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

			filtered := p.filterEvent(handler, event)
			assert.Equal(t, tc.expected, filtered)
		})
	}
}
