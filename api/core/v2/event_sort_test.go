package v2

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func eventsToId(event []*Event) []string {
	ids := make([]string, 0, len(event))
	for _, event := range event {
		ids = append(ids, string(event.ID))
	}
	return ids
}

func TestCmpBySeverity(t *testing.T) {
	critical := FixtureEvent("entity", "check")
	critical.Check.Status = 2 // crit
	critical.ID = []byte("critical")
	warn := FixtureEvent("entity", "check")
	warn.Check.Status = 1 // warn
	warn.ID = []byte("warn")
	unknown := FixtureEvent("entity", "check")
	unknown.Check.Status = 3 // unknown
	unknown.ID = []byte("unknown")
	ok := FixtureEvent("entity", "check")
	ok.Check.Status = 0 // ok
	ok.Check.LastOK = 42
	ok.ID = []byte("ok")

	testCases := []struct {
		name     string
		asc      bool
		input    []*Event
		expected []*Event
	}{
		{
			name:     "ascending",
			asc:      true,
			input:    []*Event{warn, unknown, critical, ok},
			expected: []*Event{ok, unknown, warn, critical},
		},
		{
			name:     "ascending - already sorted",
			asc:      true,
			input:    []*Event{ok, unknown, warn, critical},
			expected: []*Event{ok, unknown, warn, critical},
		},
		{
			name:     "ascending - reversed",
			asc:      true,
			input:    []*Event{critical, warn, unknown, ok},
			expected: []*Event{ok, unknown, warn, critical},
		},
		{
			name:     "descending",
			asc:      false,
			input:    []*Event{warn, unknown, critical, ok},
			expected: []*Event{critical, warn, unknown, ok},
		},
		{
			name:     "descending - already sorted",
			asc:      false,
			input:    []*Event{critical, warn, unknown, ok},
			expected: []*Event{critical, warn, unknown, ok},
		},
		{
			name:     "descending - reversed",
			asc:      false,
			input:    []*Event{ok, unknown, warn, critical},
			expected: []*Event{critical, warn, unknown, ok},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			sorter := &eventSorter{testCase.input, createCmpEvents(cmpBySeverity(testCase.asc))}
			sort.Sort(sorter)

			assert.Equal(t, eventsToId(testCase.expected), eventsToId(testCase.input))
		})
	}
}

func TestCmpByState(t *testing.T) {
	passing1 := FixtureEvent("entity", "check")
	passing1.Check.State = EventPassingState
	passing1.ID = []byte("passing1")
	passing2 := FixtureEvent("entity", "check")
	passing2.Check.State = EventPassingState
	passing2.ID = []byte("passing2")
	failing1 := FixtureEvent("entity", "check")
	failing1.Check.State = EventFailingState
	failing1.ID = []byte("failing1")
	failing2 := FixtureEvent("entity", "check")
	failing2.Check.State = EventFailingState
	failing2.ID = []byte("failing2")
	flapping1 := FixtureEvent("entity", "check")
	flapping1.Check.State = EventFlappingState
	flapping1.ID = []byte("flapping1")
	flapping2 := FixtureEvent("entity", "check")
	flapping2.Check.State = EventFlappingState
	flapping2.ID = []byte("flapping2")

	testCases := []struct {
		name     string
		asc      bool
		event1   *Event
		event2   *Event
		expected int
	}{
		{"ascending - passing - passing", true, passing1, passing2, 0},
		{"descending - passing - passing", false, passing1, passing2, 0},
		{"ascending - flapping - flapping", true, flapping1, flapping2, 0},
		{"descending - flapping - flapping", false, flapping1, flapping2, 0},
		{"ascending - failing - failing", true, failing1, failing2, 0},
		{"descending - failing - failing", false, failing1, failing2, 0},
		{"ascending - passing - failing", true, passing1, failing1, 1},
		{"descending - passing - failing", false, passing1, failing1, -1},
		{"ascending - passing - flapping", true, passing1, flapping1, 1},
		{"descending - passing - flapping", false, passing1, flapping1, -1},
		{"ascending - failing - flapping", true, flapping1, failing1, 0},
		{"descending - failing - flapping", false, failing1, flapping1, 0},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, cmpByState(testCase.asc)(testCase.event1, testCase.event2))
		})
	}
}

func TestCmpByLastOk(t *testing.T) {
	time1 := time.Now().Unix()
	time2 := time1 + 100
	time3 := time2 + 100
	time4 := time3 + 100

	lowTimeNoCheck := FixtureEvent("entity", "check")
	lowTimeNoCheck.Timestamp = time3
	lowTimeNoCheck.Check = nil
	lowTimeNoCheck.ID = []byte("lowTimeNoCheck")
	highTimeNoCheck := FixtureEvent("entity", "check")
	highTimeNoCheck.Timestamp = time4
	highTimeNoCheck.Check = nil
	highTimeNoCheck.ID = []byte("highTimeNoCheck")
	lowTimeLowCheck := FixtureEvent("entity", "check")
	lowTimeLowCheck.Timestamp = time3
	lowTimeLowCheck.Check.LastOK = time1
	lowTimeLowCheck.ID = []byte("lowTimeLowCheck")
	lowTimeHighCheck := FixtureEvent("entity", "check")
	lowTimeHighCheck.Timestamp = time3
	lowTimeHighCheck.Check.LastOK = time2
	lowTimeHighCheck.ID = []byte("lowTimeHighCheck")

	testCases := []struct {
		name     string
		asc      bool
		event1   *Event
		event2   *Event
		expected int
	}{
		{"ascending - no check vs no check", true, lowTimeNoCheck, highTimeNoCheck, 1},
		{"descending - no check vs no check", false, lowTimeNoCheck, highTimeNoCheck, -1},
		{"ascending - check vs no check", true, lowTimeLowCheck, highTimeNoCheck, 1},
		{"descending - check vs no check", false, lowTimeLowCheck, highTimeNoCheck, -1},
		{"ascending - check vs check", true, lowTimeLowCheck, lowTimeHighCheck, 1},
		{"descending - check vs check", false, lowTimeLowCheck, lowTimeHighCheck, -1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, cmpByLastOk(testCase.asc)(testCase.event1, testCase.event2))
		})
	}
}

func TestEventsBySeverity(t *testing.T) {
	critical := FixtureEvent("entity", "check")
	critical.Check.Status = 2 // crit
	critical.ID = []byte("critical")
	warn := FixtureEvent("entity", "check")
	warn.Check.Status = 1 // warn
	warn.ID = []byte("warn")
	unknown := FixtureEvent("entity", "check")
	unknown.Check.Status = 3 // unknown
	unknown.ID = []byte("unknown")
	ok := FixtureEvent("entity", "check")
	ok.Check.Status = 0 // ok
	ok.Check.LastOK = 42
	ok.ID = []byte("ok")
	okOlder := FixtureEvent("entity", "check")
	okOlder.Check.Status = 0 // ok
	okOlder.Check.LastOK = 7
	okOlder.ID = []byte("okOlder")
	okOlderDiff := FixtureEvent("entity", "check")
	okOlderDiff.Check.Status = 0 // ok
	okOlderDiff.Check.LastOK = 7
	okOlderDiff.Entity.Name = "zzz"
	okOlderDiff.ID = []byte("okOlderDiff")
	noCheck := FixtureEvent("entity", "check")
	noCheck.Check = nil
	noCheck.ID = []byte("noCheck")

	testCases := []struct {
		name     string
		asc      bool
		input    []*Event
		expected []*Event
	}{
		{
			name:     "Sorts by severity",
			asc:      false,
			input:    []*Event{ok, warn, unknown, noCheck, okOlderDiff, okOlder, critical},
			expected: []*Event{critical, warn, unknown, ok, okOlder, okOlderDiff, noCheck},
		},
		{
			name:     "Fallback to lastOK when severity is same",
			asc:      false,
			input:    []*Event{okOlder, ok, okOlder},
			expected: []*Event{ok, okOlder, okOlder},
		},
		{
			name:     "Fallback to entity name when severity is same",
			asc:      false,
			input:    []*Event{okOlderDiff, okOlder, ok, okOlder},
			expected: []*Event{ok, okOlder, okOlder, okOlderDiff},
		},
		{
			name:     "Events w/o a check are sorted to end",
			asc:      false,
			input:    []*Event{critical, noCheck, ok},
			expected: []*Event{critical, ok, noCheck},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(EventsBySeverity(tc.input, false))
			assert.EqualValues(t, eventsToId(tc.expected), eventsToId(tc.input))
		})
	}
}
func TestEventsByLastOk(t *testing.T) {
	incident := FixtureEvent("zeta", "check")
	incident.Check.State = EventFailingState
	incidentNewer := FixtureEvent("zeta", "check")
	incidentNewer.Check.State = EventFlappingState
	incidentNewer.Check.LastOK = 1
	ok := FixtureEvent("zeta", "check")
	ok.Check.State = EventPassingState
	okNewer := FixtureEvent("zeta", "check")
	okNewer.Check.State = EventPassingState
	okNewer.Check.LastOK = 1
	okDiffEntity := FixtureEvent("abba", "check")
	okDiffEntity.Check.State = EventPassingState
	okDiffCheck := FixtureEvent("abba", "0bba")
	okDiffCheck.Check.State = EventPassingState
	testCases := []struct {
		name     string
		input    []*Event
		expected []*Event
	}{
		{
			name:     "sort by lastOK",
			input:    []*Event{ok, okNewer, incidentNewer, incident},
			expected: []*Event{incidentNewer, incident, okNewer, ok},
		},
		{
			name:     "non-passing are sorted to the top",
			input:    []*Event{okNewer, incidentNewer, ok, incident},
			expected: []*Event{incidentNewer, incident, okNewer, ok},
		},
		{
			name:     "fallback to entity & check name when severity is same",
			input:    []*Event{ok, okNewer, okDiffCheck, okDiffEntity},
			expected: []*Event{okNewer, okDiffCheck, okDiffEntity, ok},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(EventsByLastOk(tc.input, false))
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
func TestEventsByEntity(t *testing.T) {
	a1 := FixtureEvent("a", "a")
	a2 := FixtureEvent("a", "b")
	b1 := FixtureEvent("b", "a")
	b2 := FixtureEvent("b", "b")
	testCases := []struct {
		name      string
		inEvents  []*Event
		ascending bool
		expected  []*Event
	}{
		{
			name:      "Sorts ascending",
			ascending: true,
			inEvents:  []*Event{b1, a2, a1, b2},
			expected:  []*Event{a1, a2, b1, b2},
		},
		{
			name:      "Sorts descending",
			ascending: false,
			inEvents:  []*Event{b1, a2, a1, b2},
			expected:  []*Event{b2, b1, a2, a1},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(EventsByEntityName(tc.inEvents, tc.ascending))
			assert.EqualValues(t, tc.expected, tc.inEvents)
		})
	}
}
