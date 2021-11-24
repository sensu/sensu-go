package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
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
		return path.Join(URLPrefix, "namespaces", url.PathEscape(e.Entity.Namespace), EventsResource, url.PathEscape(e.Entity.Name))
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

	for _, pipeline := range e.Pipelines {
		if err := e.validatePipelineReference(pipeline); err != nil {
			return errors.New("pipeline reference is invalid: " + err.Error())
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

// HasCheckHandlers determines if an event has one or more check handlers.
func (e *Event) HasCheckHandlers() bool {
	if e.HasCheck() && len(e.Check.Handlers) > 0 {
		return true
	}
	return false
}

// HasMetrics determines if an event has metric data.
func (e *Event) HasMetrics() bool {
	return e.Metrics != nil
}

// HasMetricHandlers determines if an event has one or more metric handlers.
func (e *Event) HasMetricHandlers() bool {
	if e.HasMetrics() && len(e.Metrics.Handlers) > 0 {
		return true
	}
	return false
}

// HasHandlers determines if an event has one or more check or metric handlers.
func (e *Event) HasHandlers() bool {
	if e.HasCheckHandlers() || e.HasMetricHandlers() {
		return true
	}
	return false
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

// validatePipelineReference validates that a resource reference is capable of
// acting as a pipeline.
func (e *Event) validatePipelineReference(ref *ResourceReference) error {
	switch ref.APIVersion {
	case "core/v2":
		switch ref.Type {
		case "Pipeline":
			return nil
		}
	}
	return fmt.Errorf("resource type not capable of acting as a pipeline: %s.%s", ref.APIVersion, ref.Type)
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
	if entry.Begin > now {
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
	fields := map[string]string{
		"event.name":                 resource.ObjectMeta.Name,
		"event.namespace":            resource.ObjectMeta.Namespace,
		"event.is_silenced":          isSilenced(resource),
		"event.check.is_silenced":    isSilenced(resource),
		"event.check.name":           resource.Check.Name,
		"event.check.handlers":       strings.Join(resource.Check.Handlers, ","),
		"event.check.publish":        strconv.FormatBool(resource.Check.Publish),
		"event.check.round_robin":    strconv.FormatBool(resource.Check.RoundRobin),
		"event.check.runtime_assets": strings.Join(resource.Check.RuntimeAssets, ","),
		"event.check.state":          resource.Check.State,
		"event.check.status":         strconv.Itoa(int(resource.Check.Status)),
		"event.check.subscriptions":  strings.Join(resource.Check.Subscriptions, ","),
		"event.entity.deregister":    strconv.FormatBool(resource.Entity.Deregister),
		"event.entity.name":          resource.Entity.ObjectMeta.Name,
		"event.entity.entity_class":  resource.Entity.EntityClass,
		"event.entity.subscriptions": strings.Join(resource.Entity.Subscriptions, ","),
	}
	stringsutil.MergeMapWithPrefix(fields, resource.ObjectMeta.Labels, "event.labels.")
	stringsutil.MergeMapWithPrefix(fields, resource.Entity.ObjectMeta.Labels, "event.labels.")
	stringsutil.MergeMapWithPrefix(fields, resource.Check.ObjectMeta.Labels, "event.labels.")
	return fields
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
	if e.Entity == nil {
		e.Entity = new(Entity)
	}
	e.Entity.Namespace = namespace
	if e.Check != nil {
		e.Check.Namespace = namespace
	}
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

// LogFields populates a map with fields containing relevant information about
// an event for logging
func (e *Event) LogFields(debug bool) map[string]interface{} {
	// Ensure the entity is present
	if e.Entity == nil {
		return map[string]interface{}{}
	}

	fields := map[string]interface{}{
		"event_id":         e.GetUUID().String(),
		"entity_name":      e.Entity.Name,
		"entity_namespace": e.Entity.Namespace,
	}

	if e.HasCheck() {
		fields["check_name"] = e.Check.Name
		fields["check_namespace"] = e.Check.Namespace
	}

	if debug {
		fields["timestamp"] = e.Timestamp
		if e.HasMetrics() {
			fields["metrics"] = e.Metrics
		}
		if e.HasCheck() {
			fields["hooks"] = e.Check.Hooks
			fields["silenced"] = e.Check.Silenced
		}
	} else {
		if e.HasMetrics() {
			count := len(e.Metrics.Points)
			fields["metric_count"] = count
			if count > 0 {
				fields["first_metric_name"] = e.Metrics.Points[0].Name
				fields["first_metric_value"] = e.Metrics.Points[0].Value
			}
		}
	}

	return fields
}

func (e Event) MarshalJSON() ([]byte, error) {
	type clone Event
	b, err := jsoniter.Marshal((*clone)(&e))
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
		idBytes, _ := jsoniter.Marshal(uid.String())
		msg["id"] = (*json.RawMessage)(&idBytes)
	}
	return jsoniter.Marshal(msg)
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
		b, _ = jsoniter.Marshal(msg)
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
