package v2

import (
	"errors"
	fmt "fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	stringsutil "github.com/sensu/sensu-go/util/strings"
)

// EventFailingState indicates failing check result status
const EventFailingState = "failing"

// EventFlappingState indicates a rapid change in check result status
const EventFlappingState = "flapping"

// EventPassingState indicates successful check result status
const EventPassingState = "passing"

// FixtureEvent returns a testing fixture for an Event object.
func FixtureEvent(entityName, checkID string) *Event {
	return &Event{
		ObjectMeta: NewObjectMeta("", "default"),
		Timestamp:  time.Now().Unix(),
		Entity:     FixtureEntity(entityName),
		Check:      FixtureCheck(checkID),
	}
}

// NewEvent creates a new Event.
func NewEvent(meta ObjectMeta) *Event {
	return &Event{ObjectMeta: meta}
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

	if e.Name != "" {
		return errors.New("events cannot be named")
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

// SynthesizeExtras implements dynamic.SynthesizeExtras
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
		cmpByUniqueComponents,
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
		cmpByUniqueComponents,
	)}
}

func cmpByUniqueComponents(a, b *Event) int {
	ai, bi := "", ""
	if a.Entity != nil {
		ai += a.Entity.Name
	}
	if a.Check != nil {
		ai += a.Check.Name
	}
	if b.Entity != nil {
		bi = b.Entity.Name
	}
	if b.Check != nil {
		bi += b.Check.Name
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
	return fmt.Sprintf("/api/core/v2/namespaces/%s/events/%s/%s", url.PathEscape(e.Entity.Namespace), url.PathEscape(e.Entity.Name), url.PathEscape(e.Check.Name))
}

// SilencedBy returns the subset of given silences, that silence the event.
func (e *Event) SilencedBy(entries []*Silenced) []*Silenced {
	silencedBy := make([]*Silenced, 0, len(entries))
	if !e.HasCheck() {
		return silencedBy
	}

	// Loop through every silenced entries in order to determine if it applies to
	// the given event
	for _, entry := range entries {
		if e.IsSilencedBy(entry) {
			silencedBy = append(silencedBy, entry)
		}
	}

	return silencedBy
}

// IsSilencedBy returns true if given silence will silence the event.
func (e *Event) IsSilencedBy(entry *Silenced) bool {
	if !e.HasCheck() {
		return false
	}

	// Make sure the silence has started
	now := time.Now().Unix()
	if !entry.StartSilence(now) {
		return false
	}

	// Is this event silenced for all subscriptions? (e.g. *:check_cpu)
	if entry.Name == fmt.Sprintf("*:%s", e.Check.Name) {
		return true
	}

	// Is this event silenced by the entity subscription? (e.g. entity:id:*)
	if entry.Name == fmt.Sprintf("%s:*", GetEntitySubscription(e.Entity.Name)) {
		return true
	}

	// Is this event silenced for this particular entity? (e.g.
	// entity:id:check_cpu)
	if entry.Name == fmt.Sprintf("%s:%s", GetEntitySubscription(e.Entity.Name), e.Check.Name) {
		return true
	}

	for _, subscription := range e.Check.Subscriptions {
		// Make sure the entity is subscribed to this specific subscription
		if !stringsutil.InArray(subscription, e.Entity.Subscriptions) {
			continue
		}

		// Is this event silenced by one of the check subscription? (e.g.
		// load-balancer:*)
		if entry.Name == fmt.Sprintf("%s:*", subscription) {
			return true
		}

		// Is this event silenced by one of the check subscription for this
		// particular check? (e.g. load-balancer:check_cpu)
		if entry.Name == fmt.Sprintf("%s:%s", subscription, e.Check.Name) {
			return true
		}
	}

	return false
}

// EventFields returns a set of fields that represent that resource
func EventFields(r Resource) map[string]string {
	resource := r.(*Event)
	return map[string]string{
		"event.name":                 resource.ObjectMeta.Name,
		"event.namespace":            resource.ObjectMeta.Namespace,
		"event.check.handlers":       strings.Join(resource.Check.Handlers, ","),
		"event.check.publish":        strconv.FormatBool(resource.Check.Publish),
		"event.check.round_robin":    strconv.FormatBool(resource.Check.RoundRobin),
		"event.check.runtime_assets": strings.Join(resource.Check.RuntimeAssets, ","),
		"event.check.status":         strconv.Itoa(int(resource.Check.Status)),
		"event.check.subscriptions":  strings.Join(resource.Check.Subscriptions, ","),
		"event.entity.deregister":    strconv.FormatBool(resource.Entity.Deregister),
		"event.entity.entity_class":  resource.Entity.EntityClass,
		"event.entity.subscriptions": strings.Join(resource.Entity.Subscriptions, ","),
	}
}
