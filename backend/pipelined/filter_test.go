// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestPipelinedFilter(t *testing.T) {
	p := &Pipelined{}

	testCases := []struct {
		name     string
		status   int
		metrics  *types.Metrics
		filters  []types.EventFilter
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
			name:    "Allow Filter With No Match",
			status:  1,
			metrics: nil,
			filters: []types.EventFilter{
				{
					Name:       "filter",
					Action:     types.EventFilterActionAllow,
					Statements: []string{`event.Check.Output == "bar"`},
				},
			},
			expected: true,
		},
		{
			name:    "Allow Filter With Match",
			status:  1,
			metrics: nil,
			filters: []types.EventFilter{
				{
					Name:       "filter",
					Action:     types.EventFilterActionAllow,
					Statements: []string{`event.Check.Output == "foo"`},
				},
			},
			expected: false,
		},
		{
			name:    "Deny Filter With No Match",
			status:  1,
			metrics: nil,
			filters: []types.EventFilter{
				{
					Name:       "filter",
					Action:     types.EventFilterActionDeny,
					Statements: []string{`event.Check.Output == "bar"`},
				},
			},
			expected: false,
		},
		{
			name:    "Deny Filter With Match",
			status:  1,
			metrics: nil,
			filters: []types.EventFilter{
				{
					Name:       "filter",
					Action:     types.EventFilterActionDeny,
					Statements: []string{`event.Check.Output == "foo"`},
				},
			},
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
		status   int
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
