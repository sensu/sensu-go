package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	stringsutil "github.com/sensu/sensu-go/api/core/v2/internal/stringutil"
)

const (
	// EventsResource is the name of this resource type
	EventsResource = "events"

	// EventFailingState indicates failing check result status
	EventFailingState = "failing"

	// EventFlappingState indicates a rapid change in check result status
	EventFlappingState = "flapping"

	// EventPassingState indicates successful check result status
	EventPassingState = "passing"
)

// StorePrefix returns the path prefix to this resource in the store
func (e *Event) StorePrefix() string {
	return EventsResource
}

// URIPath returns the path component of an event URI.
func (e *Event) URIPath() string {
	if !e.HasCheck() && e.Entity == nil {
		return path.Join(URLPrefix, EventsResource)
	}
	if !e.HasCheck() {
		return path.Join(URLPrefix, EventsResource, url.PathEscape(e.Entity.Name))
	}
	return path.Join(URLPrefix, "namespaces", url.PathEscape(e.Entity.Namespace), EventsResource, url.PathEscape(e.Entity.Name), url.PathEscape(e.Check.Name))
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

	if len(e.ID) > 0 {
		if _, err := uuid.FromBytes(e.ID); err != nil {
			return fmt.Errorf("event ID is invalid: %s", err)
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
	previous := e.previousOccurrence()
	return previous != nil && previous.Status != 0 && !e.IsIncident()
}

// IsSilenced determines if an event has any silenced entries
func (e *Event) IsSilenced() bool {
	if !e.HasCheck() {
		return false
	}

	return len(e.Check.Silenced) > 0
}

// IsFlappingStart determines if an event started flapping on this occurrence.
func (e *Event) IsFlappingStart() bool {
	if !e.HasCheck() {
		return false
	}

	previous := e.previousOccurrence()
	return previous != nil && !previous.Flapping && e.Check.State == EventFlappingState
}

// IsFlappingEnd determines if an event stopped flapping on this occurrence.
func (e *Event) IsFlappingEnd() bool {
	if !e.HasCheck() {
		return false
	}

	previous := e.previousOccurrence()
	return previous != nil && previous.Flapping && e.Check.State != EventFlappingState
}

// previousOccurrence returns the most recent CheckHistory item, excluding the current event.
func (e *Event) previousOccurrence() *CheckHistory {
	if !e.HasCheck() || len(e.Check.History) < 2 {
		return nil
	}
	return e.Check.previousOccurrence()
}

// SynthesizeExtras implements dynamic.SynthesizeExtras
func (e *Event) SynthesizeExtras() map[string]interface{} {
	return map[string]interface{}{
		"has_check":         e.HasCheck(),
		"has_metrics":       e.HasMetrics(),
		"is_incident":       e.IsIncident(),
		"is_resolution":     e.IsResolution(),
		"is_silenced":       e.IsSilenced(),
		"is_flapping_start": e.IsFlappingStart(),
		"is_flapping_end":   e.IsFlappingEnd(),
	}
}

// FixtureEvent returns a testing fixture for an Event object.
func FixtureEvent(entityName, checkID string) *Event {
	id := uuid.New()
	return &Event{
		ObjectMeta: NewObjectMeta("", "default"),
		Timestamp:  time.Now().Unix(),
		Entity:     FixtureEntity(entityName),
		Check:      FixtureCheck(checkID),
		ID:         id[:],
	}
}

// NewEvent creates a new Event.
func NewEvent(meta ObjectMeta) *Event {
	return &Event{ObjectMeta: meta}
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
	if !e.HasCheck() || entry == nil {
		return false
	}

	// Make sure the silence has started
	now := time.Now().Unix()
	if !entry.StartSilence(now) {
		return false
	}

	// Is this event silenced for all subscriptions? (e.g. *:check_cpu)
	// Is this event silenced by the entity subscription? (e.g. entity:id:* or entity:id:check_cpu)
	// This check being explicit here is probably not strictly necessary, as the presence
	// of the `entity:name` subscription seems to be enforced on entity creation, and
	// would be handled correctly the the subscription iteration logic below.
	if entry.Matches(e.Check.Name, GetEntitySubscription(e.Entity.Name)) {
		return true
	}

	// Alternatively, check whether any of the subscriptions of the entity match the silence.
	// It is not necessary to check the check subscriptions, because they are expected to
	// be a subset of the entity subscriptions for proxy entities, and an intersection
	// of entity and check subscriptions for non-proxy entities.
	//
	// Eg a proxy entity may have many subscriptions, but the check config that targets
	// that entity is likely to only use one of them in order to target the check at
	// a specific agent.
	//
	// Check configs for non-proxy entities on the other hand use their subscriptions to
	// both target entities and agents (as they are the same thing), and as a result
	// may have subscriptions present in the check config that are not present in the
	// entity.
	// Consider the following example:
	//    - check has subscriptions `linux` and `windows`
	//    - silence is for `windows` subscription
	//    - event is for an entity with the `linux` subscription
	// In this case, we don't want to match `linux` from the check, because the silence
	// is targeted at windows machines and the event is for a linux machine.
	//
	// To handle both of these cases correctly, we need to rely on the presence of the
	// event.check.proxy_entity_name field.
	if e.Check.ProxyEntityName != "" {
		// Proxy entity
		for _, subscription := range e.Entity.Subscriptions {
			if entry.Matches(e.Check.Name, subscription) {
				return true
			}
		}
	} else {
		// Non-proxy entity
		for _, subscription := range e.Check.Subscriptions {
			if !stringsutil.InArray(subscription, e.Entity.Subscriptions) {
				continue
			}
			if entry.Matches(e.Check.Name, subscription) {
				return true
			}
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
		"event.is_silenced":          isSilenced(resource),
		"event.check.is_silenced":    isSilenced(resource),
		"event.check.name":           resource.Check.Name,
		"event.check.handlers":       strings.Join(resource.Check.Handlers, ","),
		"event.check.publish":        strconv.FormatBool(resource.Check.Publish),
		"event.check.round_robin":    strconv.FormatBool(resource.Check.RoundRobin),
		"event.check.runtime_assets": strings.Join(resource.Check.RuntimeAssets, ","),
		"event.check.status":         strconv.Itoa(int(resource.Check.Status)),
		"event.check.subscriptions":  strings.Join(resource.Check.Subscriptions, ","),
		"event.entity.deregister":    strconv.FormatBool(resource.Entity.Deregister),
		"event.entity.name":          resource.Entity.ObjectMeta.Name,
		"event.entity.entity_class":  resource.Entity.EntityClass,
		"event.entity.subscriptions": strings.Join(resource.Entity.Subscriptions, ","),
	}
}

func isSilenced(e *Event) string {
	if len(e.Check.Silenced) > 0 {
		return "true"
	}
	return "false"
}

// SetNamespace sets the namespace of the resource.
func (e *Event) SetNamespace(namespace string) {
	e.Namespace = namespace
}

// SetObjectMeta sets the meta of the resource.
func (e *Event) SetObjectMeta(meta ObjectMeta) {
	e.ObjectMeta = meta
}

func (e *Event) RBACName() string {
	return "events"
}

// GetUUID parses a UUID from the ID bytes. It does not check errors, assuming
// that the event has already passed validation.
func (e *Event) GetUUID() uuid.UUID {
	id, _ := uuid.FromBytes(e.ID)
	return id
}

func (e Event) MarshalJSON() ([]byte, error) {
	type clone Event
	b, err := json.Marshal((*clone)(&e))
	if err != nil {
		return nil, err
	}
	if len(e.ID) == 0 {
		return b, nil
	}
	var msg map[string]*json.RawMessage
	_ = json.Unmarshal(b, &msg) // error impossible
	if len(e.ID) > 0 {
		uid, err := uuid.FromBytes(e.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid event ID: %s", err)
		}
		idBytes, _ := json.Marshal(uid.String())
		msg["id"] = (*json.RawMessage)(&idBytes)
	}
	return json.Marshal(msg)
}

func (e *Event) UnmarshalJSON(b []byte) error {
	type clone Event
	var msg map[string]*json.RawMessage
	if err := json.Unmarshal(b, &msg); err != nil {
		return err
	}
	if msg["id"] == nil {
		return json.Unmarshal(b, (*clone)(e))
	}
	var id string
	if err := json.Unmarshal(*msg["id"], &id); err != nil {
		return err
	}
	if len(id) > 0 {
		delete(msg, "id")
		b, _ = json.Marshal(msg)
	}
	if err := json.Unmarshal(b, (*clone)(e)); err != nil {
		return err
	}
	if len(id) > 0 {
		uid, err := uuid.Parse(id)
		if err != nil {
			return fmt.Errorf("invalid event id: %s", err)
		}
		e.ID = uid[:]
	}
	return nil
}
