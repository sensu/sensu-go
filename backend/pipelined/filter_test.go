// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"testing"

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
		Action:     types.EventFilterActionAllow,
		Statements: []string{`event.Check.Output == "bar"`},
	}
	allowFilterFoo := &types.EventFilter{
		Action:     types.EventFilterActionAllow,
		Statements: []string{`event.Check.Output == "foo"`},
	}
	denyFilterBar := &types.EventFilter{
		Action:     types.EventFilterActionDeny,
		Statements: []string{`event.Check.Output == "bar"`},
	}
	denyFilterFoo := &types.EventFilter{
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
			filters:  nil,
			expected: true,
		},
		{
			name:     "Incident",
			status:   1,
			metrics:  nil,
			silenced: []string{},
			filters:  nil,
			expected: false,
		},
		{
			name:     "Metrics OK Status",
			status:   0,
			metrics:  &types.Metrics{},
			silenced: []string{},
			filters:  nil,
			expected: false,
		},
		{
			name:     "Metrics Warning Status",
			status:   1,
			metrics:  &types.Metrics{},
			silenced: []string{},
			filters:  nil,
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
			name:     "Silenced With Metrics",
			status:   1,
			metrics:  &types.Metrics{},
			silenced: []string{"entity1"},
			filters:  nil,
			expected: false,
		},
		{
			name:     "Silenced Without Metrics",
			status:   1,
			metrics:  nil,
			silenced: []string{"entity1"},
			filters:  nil,
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
			filters:  nil,
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
			filters:  nil,
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
