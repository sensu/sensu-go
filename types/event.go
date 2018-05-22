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
func FixtureEvent(entityID, checkID string) *Event {
	return &Event{
		Timestamp: time.Now().Unix(),
		Entity:    FixtureEntity(entityID),
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
		return errors.New("entity " + err.Error())
	}

	if e.HasCheck() {
		if err := e.Check.Validate(); err != nil {
			return errors.New("check " + err.Error())
		}
	}

	if e.HasMetrics() {
		if err := e.Metrics.Validate(); err != nil {
			return errors.New("metrics " + err.Error())
		}
	}

	for _, hook := range e.Hooks {
		if err := hook.Validate(); err != nil {
			return errors.New("hook " + err.Error())
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
	// was a non-zero status, therefore indicating a resolution
	isResolution := (len(e.Check.History) > 0 &&
		e.Check.History[len(e.Check.History)-1].Status != 0 &&
		!e.IsIncident())

	return isResolution
}

// IsSilenced determines if an event has any silenced entries
func (e *Event) IsSilenced() bool {
	if !e.HasCheck() {
		return false
	}

	return len(e.Check.Silenced) > 0
}

// Get implements govaluate.Parameters
func (e *Event) Get(name string) (interface{}, error) {
	switch name {
	case "Timestamp":
		return e.Timestamp, nil
	case "Entity":
		return e.Entity, nil
	case "Check":
		return e.Check, nil
	case "Metrics":
		return e.Metrics, nil
	case "HasCheck":
		return e.HasCheck(), nil
	case "HasMetrics":
		return e.HasMetrics(), nil
	case "IsIncident":
		return e.IsIncident(), nil
	case "IsResolution":
		return e.IsResolution(), nil
	case "IsSilenced":
		return e.IsSilenced(), nil
	}
	return nil, errors.New("no parameter '" + name + "' found")
}

//
// Sorting

// EventsBySeverity can be used to sort a given collection of events by check
// status and timestamp.
func EventsBySeverity(es []*Event) sort.Interface {
	return &eventSorter{es, cmpBySeverity}
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

func cmpBySeverity(a, b *Event) bool {
	ap, bp := deriveSeverity(a), deriveSeverity(b)

	// Sort events with the same exit status by timestamp
	if ap == bp {
		return a.Timestamp > b.Timestamp
	}
	return ap < bp
}

// We want the order of importance to be critical (1), warning (2), unknown (3),
// and Ok (0) so we shift the check's status. If event is not a check sort to
// very end.
func deriveSeverity(e *Event) uint32 {
	if e.HasCheck() {
		return (e.Check.Status + 3) % 4
	}
	return 4
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
	return fmt.Sprintf("/events/%s/%s", url.PathEscape(e.Entity.ID), url.PathEscape(e.Check.Name))
}
