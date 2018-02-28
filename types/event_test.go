package types

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixtureEventIsValid(t *testing.T) {
	e := FixtureEvent("entity", "check")
	assert.NotNil(t, e)
	assert.NotNil(t, e.Entity)
	assert.NotNil(t, e.Check)
}

func TestEventValidate(t *testing.T) {
	event := FixtureEvent("entity", "check")

	event.Check.Name = ""
	assert.Error(t, event.Validate())
	event.Check.Name = "check"

	event.Entity.ID = ""
	assert.Error(t, event.Validate())
	event.Entity.ID = "entity"

	hook := FixtureHook("hook")
	hook.Name = ""
	event.Hooks = append(event.Hooks, hook)
	assert.Error(t, event.Validate())
	hook.Name = "hook"

	assert.NoError(t, event.Validate())
}

func TestMarshalJSON(t *testing.T) {
	event := FixtureEvent("entity", "check")
	_, err := json.Marshal(event)
	require.NoError(t, err)
}

func TestEventHasMetrics(t *testing.T) {
	testCases := []struct {
		name     string
		metrics  *Metrics
		expected bool
	}{
		{
			name:     "No Metrics",
			metrics:  nil,
			expected: false,
		},
		{
			name:     "Metrics",
			metrics:  &Metrics{},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := &Event{
				Metrics: tc.metrics,
			}
			metrics := event.HasMetrics()
			assert.Equal(t, tc.expected, metrics)
		})
	}
}

func TestEventIsIncident(t *testing.T) {
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
			event := &Event{
				Check: &Check{
					Status: tc.status,
				},
			}
			incident := event.IsIncident()
			assert.Equal(t, tc.expected, incident)
		})
	}
}

func TestEventIsResolution(t *testing.T) {
	testCases := []struct {
		name     string
		history  []CheckHistory
		status   int32
		expected bool
	}{
		{
			name:     "check has no history",
			history:  []CheckHistory{CheckHistory{}},
			status:   0,
			expected: false,
		},
		{
			name: "check has not transitioned",
			history: []CheckHistory{
				CheckHistory{Status: 1},
				CheckHistory{Status: 0},
			},
			status:   0,
			expected: false,
		},
		{
			name: "check has just transitioned",
			history: []CheckHistory{
				CheckHistory{Status: 0},
				CheckHistory{Status: 1},
			},
			status:   0,
			expected: true,
		},
		{
			name: "check has transitioned but still an incident",
			history: []CheckHistory{
				CheckHistory{Status: 0},
				CheckHistory{Status: 2},
			},
			status:   1,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := &Event{
				Check: &Check{
					History: tc.history,
					Status:  tc.status,
				},
			}
			resolution := event.IsResolution()
			assert.Equal(t, tc.expected, resolution)
		})
	}
}

func TestEventIsSilenced(t *testing.T) {
	testCases := []struct {
		name     string
		silenced []string
		expected bool
	}{
		{
			name:     "No silenced entries",
			silenced: []string{},
			expected: false,
		},
		{
			name:     "Silenced entry",
			silenced: []string{"entity1"},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := FixtureEvent("entity1", "check1")
			event.Silenced = tc.silenced
			silenced := event.IsSilenced()
			assert.Equal(t, tc.expected, silenced)
		})
	}
}

func TestEventsBySeverity(t *testing.T) {
	critical := FixtureEvent("entity", "check")
	critical.Check.Status = 1
	warn := FixtureEvent("entity", "check")
	warn.Check.Status = 2
	unknown := FixtureEvent("entity", "check")
	unknown.Check.Status = 3
	ok := FixtureEvent("entity", "check")
	ok.Check.Status = 0
	okOlder := FixtureEvent("entity", "check")
	okOlder.Timestamp = 42
	noCheck := FixtureEvent("entity", "check")
	noCheck.Check = nil

	testCases := []struct {
		name     string
		input    []*Event
		expected []*Event
	}{
		{
			name:     "Sorts by severity",
			input:    []*Event{ok, warn, unknown, noCheck, okOlder, critical},
			expected: []*Event{critical, warn, unknown, ok, okOlder, noCheck},
		},
		{
			name:     "Fallback to timestamp when severity is same",
			input:    []*Event{okOlder, ok, okOlder},
			expected: []*Event{ok, okOlder, okOlder},
		},
		{
			name:     "Events w/o a check are sorted to end",
			input:    []*Event{critical, noCheck, ok},
			expected: []*Event{critical, ok, noCheck},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(EventsBySeverity(tc.input))
			assert.EqualValues(t, tc.expected, tc.input)
		})
	}
}
