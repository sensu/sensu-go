package v2

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

	event.Entity.Name = ""
	assert.Error(t, event.Validate())
	event.Entity.Name = "entity"

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
		status   uint32
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
		status   uint32
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
			expected: true,
		},
		{
			name: "check has just transitioned",
			history: []CheckHistory{
				CheckHistory{Status: 0},
				CheckHistory{Status: 1},
			},
			status:   0,
			expected: false,
		},
		{
			name: "check has transitioned but still an incident",
			history: []CheckHistory{
				CheckHistory{Status: 2},
				CheckHistory{Status: 1},
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
		event    *Event
		silenced []string
		expected bool
	}{
		{
			name:     "No silenced entries",
			event:    FixtureEvent("entity1", "check1"),
			silenced: []string{},
			expected: false,
		},
		{
			name:     "Silenced entry",
			event:    FixtureEvent("entity1", "check1"),
			silenced: []string{"entity1"},
			expected: true,
		},
		{
			name:     "Metric without a check",
			event:    &Event{},
			silenced: []string{"entity1"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.event.Check != nil {
				tc.event.Check.Silenced = tc.silenced
			}
			silenced := tc.event.IsSilenced()
			assert.Equal(t, tc.expected, silenced)
		})
	}
}

func TestEventsBySeverity(t *testing.T) {
	critical := FixtureEvent("entity", "check")
	critical.Check.Status = 2 // crit
	warn := FixtureEvent("entity", "check")
	warn.Check.Status = 1 // warn
	unknown := FixtureEvent("entity", "check")
	unknown.Check.Status = 3 // unknown
	ok := FixtureEvent("entity", "check")
	ok.Check.Status = 0 // ok
	ok.Check.LastOK = 42
	okOlder := FixtureEvent("entity", "check")
	okOlder.Check.Status = 0 // ok
	okOlder.Check.LastOK = 7
	okOlderDiff := FixtureEvent("entity", "check")
	okOlderDiff.Check.Status = 0 // ok
	okOlderDiff.Check.LastOK = 7
	okOlderDiff.Entity.Name = "zzz"
	noCheck := FixtureEvent("entity", "check")
	noCheck.Check = nil

	testCases := []struct {
		name     string
		input    []*Event
		expected []*Event
	}{
		{
			name:     "Sorts by severity",
			input:    []*Event{ok, warn, unknown, noCheck, okOlderDiff, okOlder, critical},
			expected: []*Event{critical, warn, unknown, ok, okOlder, okOlderDiff, noCheck},
		},
		{
			name:     "Fallback to lastOK when severity is same",
			input:    []*Event{okOlder, ok, okOlder},
			expected: []*Event{ok, okOlder, okOlder},
		},
		{
			name:     "Fallback to entity name when severity is same",
			input:    []*Event{okOlderDiff, okOlder, ok, okOlder},
			expected: []*Event{ok, okOlder, okOlder, okOlderDiff},
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

func TestEventsByLastOk(t *testing.T) {
	incident := FixtureEvent("zeta", "check")
	incident.Check.Status = 2 // crit
	incidentNewer := FixtureEvent("zeta", "check")
	incidentNewer.Check.Status = 2 // crit
	incidentNewer.Check.LastOK = 1
	ok := FixtureEvent("zeta", "check")
	ok.Check.Status = 0 // ok
	okNewer := FixtureEvent("zeta", "check")
	okNewer.Check.Status = 0 // ok
	okNewer.Check.LastOK = 1
	okDiffEntity := FixtureEvent("abba", "check")
	okDiffEntity.Check.Status = 0 // ok
	okDiffCheck := FixtureEvent("abba", "0bba")
	okDiffCheck.Check.Status = 0 // ok

	testCases := []struct {
		name     string
		input    []*Event
		expected []*Event
	}{
		{
			name:     "Sorts by lastOK",
			input:    []*Event{ok, okNewer, incidentNewer, incident},
			expected: []*Event{incidentNewer, incident, okNewer, ok},
		},
		{
			name:     "incidents are sorted to the top",
			input:    []*Event{okNewer, incidentNewer, ok, incident},
			expected: []*Event{incidentNewer, incident, okNewer, ok},
		},
		{
			name:     "Fallback to entity & check name when severity is same",
			input:    []*Event{ok, okNewer, okDiffCheck, okDiffEntity},
			expected: []*Event{okNewer, okDiffCheck, okDiffEntity, ok},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(EventsByLastOk(tc.input))
			assert.EqualValues(t, tc.expected, tc.input)
		})
	}
}

func TestEventsByTimestamp(t *testing.T) {
	old := &Event{Timestamp: 3}
	older := &Event{Timestamp: 2}
	oldest := &Event{Timestamp: 1}
	okButHow := &Event{Timestamp: 0}

	testCases := []struct {
		name     string
		inEvents []*Event
		inDir    bool
		expected []*Event
	}{
		{
			name:     "Sorts ascending",
			inDir:    false,
			inEvents: []*Event{old, okButHow, oldest, older},
			expected: []*Event{okButHow, oldest, older, old},
		},
		{
			name:     "Sorts descending",
			inDir:    true,
			inEvents: []*Event{old, okButHow, oldest, older},
			expected: []*Event{old, older, oldest, okButHow},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(EventsByTimestamp(tc.inEvents, tc.inDir))
			assert.EqualValues(t, tc.expected, tc.inEvents)
		})
	}
}
