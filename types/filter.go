package types

import (
	"errors"
)

const (
	// FilterActionAllow is an action to allow events to pass through to the pipeline
	FilterActionAllow = "allow"

	// FilterActionDeny is an action to deny events from passing through to the pipeline
	FilterActionDeny = "deny"

	// DefaultFilterAction is the default action for filters
	DefaultFilterAction = FilterActionAllow
)

var (
	// FilterAllActions is a list of actions that can be used by filters
	FilterAllActions = []string{
		FilterActionAllow,
		FilterActionDeny,
	}
)

// Filter is a filter specification.
type Filter struct {
	// Name is the unique identifier for a filter
	Name string `json:"name"`

	// Action specifies to allow/deny events to continue through the pipeline
	Action string `json:"action"`

	// Attributes is a map of event attributes to match against
	Attributes map[string]interface{} `json:"attributes"`

	// Environment indicates to which env a filter belongs to
	Environment string `json:"environment"`

	// Organization indicates to which org a filter belongs to
	Organization string `json:"organization"`
}

// TODO: FilterWhen* for later use :)

// FilterWhenAttributes is the specification for "when" attributes in a filter.
type FilterWhenAttributes struct {
	// Days is an array of FilterWhenDaysAttributes
	Days []FilterWhenDays `json:"days"`
}

// FilterWhenDays is the specification for which days to use in a "when" filter.
type FilterWhenDays struct {
	All       []FilterWhenTimeRange `json:"all"`
	Sunday    []FilterWhenTimeRange `json:"sunday"`
	Monday    []FilterWhenTimeRange `json:"monday"`
	Tuesday   []FilterWhenTimeRange `json:"tuesday"`
	Wednesday []FilterWhenTimeRange `json:"wednesday"`
	Thursday  []FilterWhenTimeRange `json:"thursday"`
	Saturday  []FilterWhenTimeRange `json:"saturday"`
}

// FilterWhenTimeRange is the specification for time ranges in a "when" filter.
type FilterWhenTimeRange struct {
	// Begin is the time which the filter should begin
	Begin string `json:"begin"`

	// End is the time which the filter should end
	End string `json:"end"`
}

// Validate returns an error if the filter does not pass validation tests.
func (f *Filter) Validate() error {
	if err := ValidateName(f.Name); err != nil {
		return errors.New("filter name " + err.Error())
	}

	if len(f.Attributes) == 0 {
		return errors.New("filter attributes must be set")
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
func (f *Filter) GetOrg() string {
	return f.Organization
}

// GetEnv refers to the organization the filter belongs to
func (f *Filter) GetEnv() string {
	return f.Environment
}

// FixtureFilter returns a Filter fixture for testing.
func FixtureFilter(name string) *Filter {
	return &Filter{
		Name:   name,
		Action: FilterActionAllow,
		Attributes: map[string]interface{}{
			"check": map[string]interface{}{
				"team": "ops",
			},
		},
		Environment:  "default",
		Organization: "default",
	}
}

// FixtureDenyFilter returns a Filter fixture for testing.
func FixtureDenyFilter(name string) *Filter {
	return &Filter{
		Name:   name,
		Action: FilterActionDeny,
		Attributes: map[string]interface{}{
			"check": map[string]interface{}{
				"team": "ops",
			},
		},
		Environment:  "default",
		Organization: "default",
	}
}
