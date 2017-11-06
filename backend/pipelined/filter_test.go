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
		metrics  *types.Metrics
		filters  []string
		expected bool
	}{
		{
			name:     "Not Incident",
			status:   0,
			metrics:  nil,
			filters:  nil,
			expected: true,
		},
		{
			name:     "Incident",
			status:   1,
			metrics:  nil,
			filters:  nil,
			expected: false,
		},
		{
			name:     "Metrics OK Status",
			status:   0,
			metrics:  &types.Metrics{},
			filters:  nil,
			expected: false,
		},
		{
			name:     "Metrics Warning Status",
			status:   1,
			metrics:  &types.Metrics{},
			filters:  nil,
			expected: false,
		},
		{
			name:     "Allow Filter With No Match",
			status:   1,
			metrics:  nil,
			filters:  []string{"allowFilterBar"},
			expected: true,
		},
		{
			name:     "Allow Filter With Match",
			status:   1,
			metrics:  nil,
			filters:  []string{"allowFilterFoo"},
			expected: false,
		},
		{
			name:     "Deny Filter With No Match",
			status:   1,
			metrics:  nil,
			filters:  []string{"denyFilterBar"},
			expected: false,
		},
		{
			name:     "Deny Filter With Match",
			status:   1,
			metrics:  nil,
			filters:  []string{"denyFilterFoo"},
			expected: true,
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
					Status: tc.status,
					Output: "foo",
				},
				Entity: &types.Entity{
					Environment:  "default",
					Organization: "default",
				},
				Metrics: tc.metrics,
			}

			filtered := p.filterEvent(handler, event)
			assert.Equal(t, tc.expected, filtered)
		})
	}
}

func TestPipelinedIsIncident(t *testing.T) {
	p := &Pipelined{}

	testCases := []struct {
		name     string
		status   int32
		expected bool
	}{
		{
			name:     "OK Status",
			status:   0,
			expected: false,
		},
		{
			name:     "Non-zero Status",
			status:   1,
			expected: true,
		},
		{
			name:     "Negative Status",
			status:   -1,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := &types.Event{
				Check: &types.Check{
					Status: tc.status,
				},
			}
			incident := p.isIncident(event)
			assert.Equal(t, tc.expected, incident)
		})
	}
}

func TestPipelinedHasMetrics(t *testing.T) {
	p := &Pipelined{}

	testCases := []struct {
		name     string
		metrics  *types.Metrics
		expected bool
	}{
		{
			name:     "Not Metrics",
			metrics:  nil,
			expected: false,
		},
		{
			name:     "Metrics",
			metrics:  &types.Metrics{},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := &types.Event{
				Metrics: tc.metrics,
			}
			metrics := p.hasMetrics(event)
			assert.Equal(t, tc.expected, metrics)
		})
	}
}
