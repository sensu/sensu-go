package v2

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/sensu/sensu-go/api/core/v2/internal/js"
	stringsutil "github.com/sensu/sensu-go/api/core/v2/internal/stringutil"
	utilstrings "github.com/sensu/sensu-go/api/core/v2/internal/stringutil"
)

const (
	// EventFiltersResource is the name of this resource type
	EventFiltersResource = "filters"

	// EventFilterActionAllow is an action to allow events to pass through to the pipeline
	EventFilterActionAllow = "allow"

	// EventFilterActionDeny is an action to deny events from passing through to the pipeline
	EventFilterActionDeny = "deny"

	// DefaultEventFilterAction is the default action for filters
	DefaultEventFilterAction = EventFilterActionAllow
)

var (
	// EventFilterAllActions is a list of actions that can be used by filters
	EventFilterAllActions = []string{
		EventFilterActionAllow,
		EventFilterActionDeny,
	}
)

// StorePrefix returns the path prefix to this resource in the store
func (f *EventFilter) StorePrefix() string {
	return "event-filters"
}

// URIPath returns the path component of an event filter URI.
func (f *EventFilter) URIPath() string {
	if f.Namespace == "" {
		return path.Join(URLPrefix, EventFiltersResource, url.PathEscape(f.Name))
	}
	return path.Join(URLPrefix, "namespaces", url.PathEscape(f.Namespace), EventFiltersResource, url.PathEscape(f.Name))
}

// Validate returns an error if the filter does not pass validation tests.
func (f *EventFilter) Validate() error {
	if err := ValidateName(f.Name); err != nil {
		return errors.New("filter name " + err.Error())
	}

	if found := utilstrings.InArray(f.Action, EventFilterAllActions); !found {
		return fmt.Errorf("action '%s' is not valid", f.Action)
	}

	if len(f.Expressions) == 0 {
		return errors.New("filter must have one or more expressions")
	}

	if err := js.ParseExpressions(f.Expressions); err != nil {
		return err
	}

	if f.Namespace == "" {
		return errors.New("namespace must be set")
	}

	return nil
}

// Update updates e with selected fields. Returns non-nil error if any of the
// selected fields are unsupported.
func (f *EventFilter) Update(from *EventFilter, fields ...string) error {
	for _, field := range fields {
		switch field {
		case "Action":
			f.Action = from.Action
		case "When":
			f.When = from.When
		case "Expressions":
			f.Expressions = append(f.Expressions[0:0], from.Expressions...)
		case "RuntimeAssets":
			f.RuntimeAssets = append(f.RuntimeAssets[0:0], from.RuntimeAssets...)
		default:
			return fmt.Errorf("unsupported field: %q", f)
		}
	}
	return nil
}

// NewEventFilter creates a new EventFilter.
func NewEventFilter(meta ObjectMeta) *EventFilter {
	return &EventFilter{ObjectMeta: meta}
}

// FixtureEventFilter returns a Filter fixture for testing.
func FixtureEventFilter(name string) *EventFilter {
	return &EventFilter{
		Action:      EventFilterActionAllow,
		Expressions: []string{"event.check.team == 'ops'"},
		ObjectMeta:  NewObjectMeta(name, "default"),
	}
}

// FixtureDenyEventFilter returns a Filter fixture for testing.
func FixtureDenyEventFilter(name string) *EventFilter {
	return &EventFilter{
		Action:      EventFilterActionDeny,
		Expressions: []string{"event.check.team == 'ops'"},
		ObjectMeta: ObjectMeta{
			Namespace: "default",
			Name:      name,
		},
	}
}

//
// Sorting
//
type cmpEventFilter func(a, b *EventFilter) bool

// SortEventFiltersByPredicate can be used to sort a given collection using a given
// predicate.
func SortEventFiltersByPredicate(ef []*EventFilter, fn cmpEventFilter) sort.Interface {
	return &eventFilterSorter{eventFilters: ef, byFn: fn}
}

// SortEventFiltersByName can be used to sort a given collection of event filter by
// their names.
func SortEventFiltersByName(ef []*EventFilter, asc bool) sort.Interface {
	if asc {
		return SortEventFiltersByPredicate(ef, func(a, b *EventFilter) bool {
			return a.Name < b.Name
		})
	}

	return SortEventFiltersByPredicate(ef, func(a, b *EventFilter) bool {
		return a.Name > b.Name
	})
}

type eventFilterSorter struct {
	eventFilters []*EventFilter
	byFn         cmpEventFilter
}

// Len implements sort.Interface.
func (s *eventFilterSorter) Len() int {
	return len(s.eventFilters)
}

// Swap implements sort.Interface.
func (s *eventFilterSorter) Swap(i, j int) {
	s.eventFilters[i], s.eventFilters[j] = s.eventFilters[j], s.eventFilters[i]
}

// Less implements sort.Interface.
func (s *eventFilterSorter) Less(i, j int) bool {
	return s.byFn(s.eventFilters[i], s.eventFilters[j])
}

// EventFilterFields returns a set of fields that represent that resource
func EventFilterFields(r Resource) map[string]string {
	resource := r.(*EventFilter)
	fields := map[string]string{
		"filter.name":           resource.ObjectMeta.Name,
		"filter.namespace":      resource.ObjectMeta.Namespace,
		"filter.action":         resource.Action,
		"filter.runtime_assets": strings.Join(resource.RuntimeAssets, ","),
	}
	stringsutil.MergeMapWithPrefix(fields, resource.ObjectMeta.Labels, "filter.labels.")
	return fields
}

// SetNamespace sets the namespace of the resource.
func (f *EventFilter) SetNamespace(namespace string) {
	f.Namespace = namespace
}

// SetObjectMeta sets the meta of the resource.
func (f *EventFilter) SetObjectMeta(meta ObjectMeta) {
	f.ObjectMeta = meta
}

func (f *EventFilter) RBACName() string {
	return "filters"
}
