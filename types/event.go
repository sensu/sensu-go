package types

import (
	"encoding/json"
	"errors"
	fmt "fmt"
	"net/url"
	"sort"
	"time"
)

// EventFailingState indicates failing check result status
const EventFailingState = "failing"

// EventFlappingState indicates a rapid change in check result status
const EventFlappingState = "flapping"

// EventPassingState indicates successful check result status
const EventPassingState = "passing"

// FixtureEvent returns a testing fixutre for an Event object.
func FixtureEvent(entityName, checkID string) *Event {
	return &Event{
		Timestamp: time.Now().Unix(),
		Entity:    FixtureEntity(entityName),
		Check:     FixtureCheck(checkID),
	}
}

// UnmarshalJSON ...
func (e *Event) UnmarshalJSON(b []byte) error {
	// HACK HACK HACK
	// This method is a compatibility shim that should be removed
	// when we remove Silenceds and Hooks from Events
	type Evt Event
	var evt Evt
	if err := json.Unmarshal(b, &evt); err != nil {
		return err
	}
	if evt.Check == nil {
		*e = Event(evt)
		return nil
	}
	silenced := make(map[string]struct{})
	for _, s := range append(evt.Silenced, evt.Check.Silenced...) {
		silenced[s] = struct{}{}
	}
	newSilenced := make([]string, 0, len(silenced))
	for s := range silenced {
		newSilenced = append(newSilenced, s)
	}
	if len(newSilenced) > 0 {
		evt.Check.Silenced = newSilenced
	}
	evt.Silenced = nil

	hooks := make(map[*Hook]struct{})
	for _, h := range append(evt.Hooks, evt.Check.Hooks...) {
		hooks[h] = struct{}{}
	}
	newHooks := make([]*Hook, 0, len(hooks))
	for h := range hooks {
		newHooks = append(newHooks, h)
	}
	if len(newHooks) > 0 {
		evt.Check.Hooks = newHooks
	}
	evt.Hooks = nil

	*e = Event(evt)
	return nil
}

// Validate returns an error if the event does not pass validation tests.
func (e *Event) Validate() error {
	if e.Entity == nil {
		return errors.New("event must contain an entity")
	}

	if !e.HasCheck() && !e.HasMetrics() {
		return errors.New("event must contain a check or metrics")
	}

	if err := e.Entity.Validate(); err != nil {
		return errors.New("entity is invalid: " + err.Error())
	}

	if e.HasCheck() {
		if err := e.Check.Validate(); err != nil {
			return errors.New("check is invalid: " + err.Error())
		}
	}

	if e.HasMetrics() {
		if err := e.Metrics.Validate(); err != nil {
			return errors.New("metrics are invalid: " + err.Error())
		}
	}

	for _, hook := range e.Hooks {
		if err := hook.Validate(); err != nil {
			return errors.New("hook is invalid: " + err.Error())
		}
	}

	return nil
}

// HasCheck determines if an event has check data.
func (e *Event) HasCheck() bool {
	return e.Check != nil
}

// HasMetrics determines if an event has metric data.
func (e *Event) HasMetrics() bool {
	return e.Metrics != nil
}

// IsIncident determines if an event indicates an incident.
func (e *Event) IsIncident() bool {
	return e.HasCheck() && e.Check.Status != 0
}

// IsResolution returns true if an event has just transitionned from an incident
func (e *Event) IsResolution() bool {
	if !e.HasCheck() {
		return false
	}

	// Try to retrieve the previous status in the check history and verify if it
	// was a non-zero status, therefore indicating a resolution. The current event
	// has already been added to the check history by eventd so we must retrieve
	// the second to the last
	return (len(e.Check.History) > 1 &&
		e.Check.History[len(e.Check.History)-2].Status != 0 &&
		!e.IsIncident())
}

// IsSilenced determines if an event has any silenced entries
func (e *Event) IsSilenced() bool {
	if !e.HasCheck() {
		return false
	}

	return len(e.Check.Silenced) > 0
}

// Implements dynamic.SynthesizeExtras
func (e *Event) SynthesizeExtras() map[string]interface{} {
	return map[string]interface{}{
		"has_check":     e.HasCheck(),
		"has_metrics":   e.HasMetrics(),
		"is_incident":   e.IsIncident(),
		"is_resolution": e.IsResolution(),
		"is_silenced":   e.IsSilenced(),
	}
}

//
// Sorting

// EventsBySeverity can be used to sort a given collection of events by check
// status and timestamp.
func EventsBySeverity(es []*Event) sort.Interface {
	return &eventSorter{es, createCmpEvents(
		cmpBySeverity,
		cmpByLastOk,
		cmpByEntityName,
	)}
}

// EventsByTimestamp can be used to sort a given collection of events by time it
// occurred.
func EventsByTimestamp(es []*Event, asc bool) sort.Interface {
	sorter := &eventSorter{events: es}
	if asc {
		sorter.byFn = func(a, b *Event) bool {
			return a.Timestamp > b.Timestamp
		}
	} else {
		sorter.byFn = func(a, b *Event) bool {
			return a.Timestamp < b.Timestamp
		}
	}
	return sorter
}

// EventsByLastOk can be used to sort a given collection of events by time it
// last received an OK status.
func EventsByLastOk(es []*Event) sort.Interface {
	return &eventSorter{es, createCmpEvents(
		cmpByIncident,
		cmpByLastOk,
		cmpByEntityName,
	)}
}

func cmpByEntityName(a, b *Event) int {
	ai, bi := "", ""
	if a.Entity != nil {
		ai = a.Entity.Name
	}
	if b.Entity != nil {
		bi = b.Entity.Name
	}

	if ai == bi {
		return 0
	} else if ai < bi {
		return 1
	}
	return -1
}

func cmpBySeverity(a, b *Event) int {
	ap, bp := deriveSeverity(a), deriveSeverity(b)

	// Sort events with the same exit status by timestamp
	if ap == bp {
		return 0
	} else if ap < bp {
		return 1
	}
	return -1
}

func cmpByIncident(a, b *Event) int {
	av, bv := a.IsIncident(), b.IsIncident()

	// Rank higher if incident
	if av == bv {
		return 0
	} else if av {
		return 1
	}
	return -1
}

func cmpByLastOk(a, b *Event) int {
	at, bt := a.Timestamp, b.Timestamp
	if a.HasCheck() {
		at = a.Check.LastOK
	}
	if b.HasCheck() {
		bt = b.Check.LastOK
	}

	if at == bt {
		return 0
	} else if at > bt {
		return 1
	}
	return -1
}

// Based on convention we define the order of importance as critical (2),
// warning (1), unknown (>2), and Ok (0). If event is not a check sort to
// very end.
func deriveSeverity(e *Event) int {
	if e.HasCheck() {
		switch e.Check.Status {
		case 0:
			return 3
		case 1:
			return 1
		case 2:
			return 0
		default:
			return 2
		}
	}
	return 4
}

type cmpEvents func(a, b *Event) int

func createCmpEvents(cmps ...cmpEvents) func(a, b *Event) bool {
	return func(a, b *Event) bool {
		for _, cmp := range cmps {
			st := cmp(a, b)
			if st == 0 { // if equal try the next comparitor
				continue
			}
			return st == 1
		}
		return true
	}
}

type eventSorter struct {
	events []*Event
	byFn   func(a, b *Event) bool
}

// Len implements sort.Interface.
func (s *eventSorter) Len() int {
	return len(s.events)
}

// Swap implements sort.Interface.
func (s *eventSorter) Swap(i, j int) {
	s.events[i], s.events[j] = s.events[j], s.events[i]
}

// Less implements sort.Interface.
func (s *eventSorter) Less(i, j int) bool {
	return s.byFn(s.events[i], s.events[j])
}

// URIPath returns the path component of a Event URI.
func (e *Event) URIPath() string {
	if !e.HasCheck() {
		return ""
	}
	return fmt.Sprintf("/events/%s/%s", url.PathEscape(e.Entity.Name), url.PathEscape(e.Check.Name))
}
