package types

import (
	"errors"
)

const (
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

// EventFilter is a filter specification.
type EventFilter struct {
	// Name is the unique identifier for a filter
	Name string `json:"name"`

	// Action specifies to allow/deny events to continue through the pipeline
	Action string `json:"action"`

	// Statements is an array of boolean expressions that are &&'d together
	// to determine if the event matches this filter.
	Statements []string `json:"statements"`

	// Environment indicates to which env a filter belongs to
	Environment string `json:"environment"`

	// Organization indicates to which org a filter belongs to
	Organization string `json:"organization"`
}

// TODO: EventFilterWhen* for later use :)

// EventFilterWhenAttributes is the specification for "when" attributes in a filter.
type EventFilterWhenAttributes struct {
	// Days is an array of EventFilterWhenDaysAttributes
	Days []EventFilterWhenDays `json:"days"`
}

// EventFilterWhenDays is the specification for which days to use in a "when" filter.
type EventFilterWhenDays struct {
	All       []EventFilterWhenTimeRange `json:"all"`
	Sunday    []EventFilterWhenTimeRange `json:"sunday"`
	Monday    []EventFilterWhenTimeRange `json:"monday"`
	Tuesday   []EventFilterWhenTimeRange `json:"tuesday"`
	Wednesday []EventFilterWhenTimeRange `json:"wednesday"`
	Thursday  []EventFilterWhenTimeRange `json:"thursday"`
	Saturday  []EventFilterWhenTimeRange `json:"saturday"`
}

// EventFilterWhenTimeRange is the specification for time ranges in a "when" filter.
type EventFilterWhenTimeRange struct {
	// Begin is the time which the filter should begin
	Begin string `json:"begin"`

	// End is the time which the filter should end
	End string `json:"end"`
}

// Validate returns an error if the filter does not pass validation tests.
func (f *EventFilter) Validate() error {
	if err := ValidateName(f.Name); err != nil {
		return errors.New("filter name " + err.Error())
	}

	if len(f.Statements) == 0 {
		return errors.New("filter must have one or more statements")
	}

	if f.Environment == "" {
		return errors.New("environment must be set")
	}

	if f.Organization == "" {
		return errors.New("organization must be set")
	}

	return nil
}

// GetOrg refers to the organization the filter belongs to
func (f *EventFilter) GetOrg() string {
	return f.Organization
}

// GetEnv refers to the organization the filter belongs to
func (f *EventFilter) GetEnv() string {
	return f.Environment
}

// FixtureEventFilter returns a Filter fixture for testing.
func FixtureEventFilter(name string) *EventFilter {
	return &EventFilter{
		Name:         name,
		Action:       EventFilterActionAllow,
		Statements:   []string{"event.Check.Team == 'ops'"},
		Environment:  "default",
		Organization: "default",
	}
}

// FixtureDenyEventFilter returns a Filter fixture for testing.
func FixtureDenyEventFilter(name string) *EventFilter {
	return &EventFilter{
		Name:         name,
		Action:       EventFilterActionDeny,
		Statements:   []string{"event.Check.Team == 'ops'"},
		Environment:  "default",
		Organization: "default",
	}
}
