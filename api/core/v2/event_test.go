package v2

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
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

func TestEventValidateNoTimestamp(t *testing.T) {
	// Events without a timestamp are valid
	event := FixtureEvent("entity", "check")
	event.Timestamp = 0
	if err := event.Validate(); err != nil {
		t.Fatal(err)
	}
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

func TestEventIsFlappingStart(t *testing.T) {
	testCases := []struct {
		name     string
		history  []CheckHistory
		state    string
		expected bool
	}{
		{
			name:     "check has no history",
			history:  []CheckHistory{CheckHistory{}},
			state:    EventPassingState,
			expected: false,
		},
		{
			name: "check was not flapping previously, nor is now",
			history: []CheckHistory{
				CheckHistory{Flapping: false},
				CheckHistory{Flapping: false},
			},
			state:    EventPassingState,
			expected: false,
		},
		{
			name: "check is already flapping",
			history: []CheckHistory{
				CheckHistory{Flapping: true},
				CheckHistory{Flapping: true},
			},
			state:    EventFlappingState,
			expected: false,
		},
		{
			name: "check was not previously flapping",
			history: []CheckHistory{
				CheckHistory{Flapping: false},
				CheckHistory{Flapping: true},
			},
			state:    EventFlappingState,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := &Event{
				Check: &Check{
					History: tc.history,
					State:   tc.state,
				},
			}
			assert.Equal(t, tc.expected, event.IsFlappingStart())
		})
	}
}

func TestEventIsFlappingEnd(t *testing.T) {
	testCases := []struct {
		name     string
		history  []CheckHistory
		state    string
		expected bool
	}{
		{
			name:     "check has no history",
			history:  []CheckHistory{CheckHistory{}},
			state:    EventPassingState,
			expected: false,
		},
		{
			name: "check was not flapping previously, nor is now",
			history: []CheckHistory{
				CheckHistory{Flapping: false},
				CheckHistory{Flapping: false},
			},
			state:    EventPassingState,
			expected: false,
		},
		{
			name: "check is already flapping",
			history: []CheckHistory{
				CheckHistory{Flapping: true},
				CheckHistory{Flapping: true},
			},
			state:    EventFlappingState,
			expected: false,
		},
		{
			name: "check was previously flapping but now is in OK state",
			history: []CheckHistory{
				CheckHistory{Flapping: true},
				CheckHistory{Flapping: false},
			},
			state:    EventPassingState,
			expected: true,
		},
		{
			name: "check was previously flapping but now is in failing state",
			history: []CheckHistory{
				CheckHistory{Flapping: true},
				CheckHistory{Flapping: false},
			},
			state:    EventFailingState,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := &Event{
				Check: &Check{
					History: tc.history,
					State:   tc.state,
				},
			}
			assert.Equal(t, tc.expected, event.IsFlappingEnd())
		})
	}
}

func TestEventIsSilencedBy(t *testing.T) {
	testCases := []struct {
		name     string
		event    *Event
		silenced *Silenced
		expected bool
	}{
		{
			name:     "nil silenced",
			event:    FixtureEvent("entity1", "check1"),
			silenced: nil,
			expected: false,
		},
		{
			name:     "Metric without a check",
			event:    &Event{},
			silenced: nil,
			expected: false,
		},
		{
			name:     "Check provided and doesn't match, no subscription",
			event:    FixtureEvent("entity1", "check1"),
			silenced: FixtureSilenced("*:check2"),
			expected: false,
		},
		{
			name:     "Check provided and matches, no subscription",
			event:    FixtureEvent("entity1", "check1"),
			silenced: FixtureSilenced("*:check1"),
			expected: true,
		},
		{
			name:     "Subscription provided and doesn't match, no check",
			event:    FixtureEvent("entity1", "check1"),
			silenced: FixtureSilenced("entity:entity2:*"),
			expected: false,
		},
		{
			name:     "Subscription provided and matches, no check",
			event:    FixtureEvent("entity1", "check1"),
			silenced: FixtureSilenced("entity:entity1:*"),
			expected: true,
		},
		{
			name:     "Check provided and doesn't match, subscription matches",
			event:    FixtureEvent("entity1", "check1"),
			silenced: FixtureSilenced("entity:entity1:check2"),
			expected: false,
		},
		{
			name:     "Check provided and matches, subscription doesn't match",
			event:    FixtureEvent("entity1", "check1"),
			silenced: FixtureSilenced("entity:entity2:check1"),
			expected: false,
		},
		{
			name:     "Check and subscription both provided and match",
			event:    FixtureEvent("entity1", "check1"),
			silenced: FixtureSilenced("entity:entity1:check1"),
			expected: true,
		},
		{
			name:     "Subscription provided and doesn't match, check matches",
			event:    FixtureEvent("entity1", "check1"),
			silenced: FixtureSilenced("entity:entity2:check1"),
			expected: false,
		},
		{
			name:     "Subscription provided and matches, check doesn't match",
			event:    FixtureEvent("entity1", "check1"),
			silenced: FixtureSilenced("entity:entity1:check2"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.event.IsSilencedBy(tc.silenced))
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

func TestSilencedBy(t *testing.T) {
	testCases := []struct {
		name            string
		event           *Event
		entries         []*Silenced
		expectedEntries []*Silenced
	}{
		{
			name:            "no entries",
			event:           FixtureEvent("foo", "check_cpu"),
			entries:         []*Silenced{},
			expectedEntries: []*Silenced{},
		},
		{
			name:  "not silenced",
			event: FixtureEvent("foo", "check_cpu"),
			entries: []*Silenced{
				FixtureSilenced("entity:foo:check_mem"),
				FixtureSilenced("entity:bar:*"),
				FixtureSilenced("foo:check_cpu"),
				FixtureSilenced("foo:*"),
				FixtureSilenced("*:check_mem"),
			},
			expectedEntries: []*Silenced{},
		},
		{
			name:  "silenced by check",
			event: FixtureEvent("foo", "check_cpu"),
			entries: []*Silenced{
				FixtureSilenced("*:check_cpu"),
			},
			expectedEntries: []*Silenced{
				FixtureSilenced("*:check_cpu"),
			},
		},
		{
			name:  "silenced by entity subscription",
			event: FixtureEvent("foo", "check_cpu"),
			entries: []*Silenced{
				FixtureSilenced("entity:foo:*"),
			},
			expectedEntries: []*Silenced{
				FixtureSilenced("entity:foo:*"),
			},
		},
		{
			name:  "silenced by entity's check subscription",
			event: FixtureEvent("foo", "check_cpu"),
			entries: []*Silenced{
				FixtureSilenced("entity:foo:check_cpu"),
			},
			expectedEntries: []*Silenced{
				FixtureSilenced("entity:foo:check_cpu"),
			},
		},
		{
			name:  "silenced by check subscription",
			event: FixtureEvent("foo", "check_cpu"), // has a linux subscription
			entries: []*Silenced{
				FixtureSilenced("linux:*"),
			},
			expectedEntries: []*Silenced{
				FixtureSilenced("linux:*"),
			},
		},
		{
			name:  "silenced by subscription with check",
			event: FixtureEvent("foo", "check_cpu"), // has a linux subscription
			entries: []*Silenced{
				FixtureSilenced("linux:check_cpu"),
			},
			expectedEntries: []*Silenced{
				FixtureSilenced("linux:check_cpu"),
			},
		},
		{
			name:  "silenced by multiple entries",
			event: FixtureEvent("foo", "check_cpu"), // has a linux subscription
			entries: []*Silenced{
				FixtureSilenced("entity:foo:*"),
				FixtureSilenced("linux:check_cpu"),
			},
			expectedEntries: []*Silenced{
				FixtureSilenced("entity:foo:*"),
				FixtureSilenced("linux:check_cpu"),
			},
		},
		{
			name: "not silenced, silenced & client don't have a common subscription",
			event: &Event{
				Check: &Check{
					ObjectMeta: ObjectMeta{
						Name: "check_cpu",
					},
					Subscriptions: []string{"linux", "windows"},
				},
				Entity: &Entity{
					ObjectMeta: ObjectMeta{
						Name: "foo",
					},
					Subscriptions: []string{"linux"},
				},
			},
			entries: []*Silenced{
				FixtureSilenced("windows:check_cpu"),
			},
			expectedEntries: []*Silenced{},
		},
		{
			name: "silenced, silenced & client do have a common subscription",
			event: &Event{
				Check: &Check{
					ObjectMeta: ObjectMeta{
						Name: "check_cpu",
					},
					Subscriptions: []string{"linux", "windows"},
				},
				Entity: &Entity{
					ObjectMeta: ObjectMeta{
						Name: "foo",
					},
					Subscriptions: []string{"linux"},
				},
			},
			entries: []*Silenced{
				FixtureSilenced("linux:check_cpu"),
			},
			expectedEntries: []*Silenced{
				FixtureSilenced("linux:check_cpu"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.event.SilencedBy(tc.entries)
			assert.EqualValues(t, tc.expectedEntries, result)
		})
	}
}

func TestIsSilencedBy(t *testing.T) {
	testCases := []struct {
		name           string
		event          *Event
		silence        *Silenced
		expectedResult bool
	}{
		{
			name:  "silence has not started",
			event: FixtureEvent("foo", "check_cpu"),
			silence: &Silenced{
				ObjectMeta: ObjectMeta{
					Name: "*:check_cpu",
				},
				Begin: time.Now().Add(1 * time.Hour).Unix(),
			},
			expectedResult: false,
		},
		{
			name:           "check matches w/ wildcard subscription",
			event:          FixtureEvent("foo", "check_cpu"),
			silence:        FixtureSilenced("*:check_cpu"),
			expectedResult: true,
		},
		{
			name:           "entity subscription matches w/ wildcard check",
			event:          FixtureEvent("foo", "check_cpu"),
			silence:        FixtureSilenced("entity:foo:*"),
			expectedResult: true,
		},
		{
			name:           "entity subscription and check match",
			event:          FixtureEvent("foo", "check_cpu"),
			silence:        FixtureSilenced("entity:foo:check_cpu"),
			expectedResult: true,
		},
		{
			name: "subscription matches",
			event: &Event{
				Check: &Check{
					ObjectMeta: ObjectMeta{
						Name: "check_cpu",
					},
					Subscriptions: []string{"unix"},
				},
				Entity: &Entity{
					ObjectMeta: ObjectMeta{
						Name: "foo",
					},
					Subscriptions: []string{"unix"},
				},
			},
			silence:        FixtureSilenced("unix:check_cpu"),
			expectedResult: true,
		},
		{
			name: "subscription does not match",
			event: &Event{
				Check: &Check{
					ObjectMeta: ObjectMeta{
						Name: "check_cpu",
					},
					Subscriptions: []string{"unix"},
				},
				Entity: &Entity{
					ObjectMeta: ObjectMeta{
						Name: "foo",
					},
					Subscriptions: []string{"unix"},
				},
			},
			silence:        FixtureSilenced("windows:check_cpu"),
			expectedResult: false,
		},
		{
			name: "entity subscription doesn't match",
			event: &Event{
				Check: &Check{
					ObjectMeta: ObjectMeta{
						Name: "check_cpu",
					},
					Subscriptions: []string{"unix"},
				},
				Entity: &Entity{
					ObjectMeta: ObjectMeta{
						Name: "foo",
					},
					Subscriptions: []string{"windows"},
				},
			},
			silence: &Silenced{
				Subscription: "check",
				Check:        "check_cpu",
			},
			expectedResult: false,
		},
		{
			name:           "check does not match",
			event:          FixtureEvent("foo", "check_mem"),
			silence:        FixtureSilenced("*:check_cpu"),
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.event.IsSilencedBy(tc.silence)
			assert.EqualValues(t, tc.expectedResult, result)
		})
	}
}

func fixtureNoID() *Event {
	e := FixtureEvent("foo", "bar")
	e.ID = nil
	return e
}

func fixtureBadID() *Event {
	e := FixtureEvent("foo", "bar")
	e.ID = []byte("not a uuid")
	return e
}

func TestMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		Name         string
		Event        *Event
		MarshalError bool
	}{
		{
			Name:  "event with no ID",
			Event: fixtureNoID(),
		},
		{
			Name:  "event with ID",
			Event: FixtureEvent("foo", "bar"),
		},
		{
			Name:         "event with invalid ID",
			Event:        fixtureBadID(),
			MarshalError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			b, err := json.Marshal(test.Event)
			if test.MarshalError {
				if err == nil {
					t.Fatal("expected non-nil error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var e Event
			if err := json.Unmarshal(b, &e); err != nil {
				t.Fatal(err)
			}
			if err := e.Validate(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestUnmarshalID(t *testing.T) {
	tests := []struct {
		Name           string
		Data           string
		UnmarshalError bool
	}{
		{
			Name: "no id",
			Data: `{}`,
		},
		{
			Name: "has id",
			Data: fmt.Sprintf(`{"id": %q}`, uuid.NameSpaceDNS),
		},
		{
			Name:           "invalid id",
			Data:           `{"id": "not a uuid"}`,
			UnmarshalError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var e Event
			err := json.Unmarshal([]byte(test.Data), &e)
			if test.UnmarshalError {
				if err == nil {
					t.Fatal("expected non-nil error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEventFields(t *testing.T) {
	tests := []struct {
		name    string
		args    Resource
		wantKey string
		want    string
	}{
		{
			name:    "exposes entity.name",
			args:    FixtureEvent("frank", "reynolds"),
			wantKey: "event.entity.name",
			want:    "frank",
		},
		{
			name:    "exposes check.name",
			args:    FixtureEvent("frank", "reynolds"),
			wantKey: "event.check.name",
			want:    "reynolds",
		},
		{
			name: "exposes check labels",
			args: &Event{
				Check:  &Check{ObjectMeta: ObjectMeta{Labels: map[string]string{"src": "bonsai"}}},
				Entity: &Entity{},
			},
			wantKey: "event.labels.src",
			want:    "bonsai",
		},
		{
			name: "exposes entity labels",
			args: &Event{
				Check:  &Check{},
				Entity: &Entity{ObjectMeta: ObjectMeta{Labels: map[string]string{"region": "philadelphia"}}},
			},
			wantKey: "event.labels.region",
			want:    "philadelphia",
		},
		{
			name: "check labels take precendence",
			args: &Event{
				Check:  &Check{ObjectMeta: ObjectMeta{Labels: map[string]string{"dupe": "check"}}},
				Entity: &Entity{ObjectMeta: ObjectMeta{Labels: map[string]string{"dupe": "entity"}}},
			},
			wantKey: "event.labels.dupe",
			want:    "check",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EventFields(tt.args)
			if !reflect.DeepEqual(got[tt.wantKey], tt.want) {
				t.Errorf("EventFields() = got[%s] %v, want[%s] %v", tt.wantKey, got[tt.wantKey], tt.wantKey, tt.want)
			}
		})
	}
}
