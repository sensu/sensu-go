package types

import (
	"errors"
	"fmt"

	"github.com/Knetic/govaluate"
	utilstrings "github.com/sensu/sensu-go/util/strings"
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

// Validate returns an error if the filter does not pass validation tests.
func (f *EventFilter) Validate() error {
	if err := ValidateName(f.Name); err != nil {
		return errors.New("filter name " + err.Error())
	}

	if found := utilstrings.InArray(f.Action, EventFilterAllActions); !found {
		return fmt.Errorf("action '%s' is not valid", f.Action)
	}

	if len(f.Statements) == 0 {
		return errors.New("filter must have one or more statements")
	}

	if err := validateStatements(f.Statements); err != nil {
		return err
	}

	if f.Environment == "" {
		return errors.New("environment must be set")
	}

	if f.Organization == "" {
		return errors.New("organization must be set")
	}

	return nil
}

// validateStatements ensure that the given statements can be parsed
// successfully and that it does not contain any modifier tokens.
func validateStatements(statements []string) error {
	for _, statement := range statements {
		exp, err := govaluate.NewEvaluableExpression(statement)
		if err != nil {
			return fmt.Errorf("invalid statement '%s': %s", statement, err.Error())
		}

		// Do not allow modifier tokens (eg. +, -, /, *, **, &, etc.)
		for _, token := range exp.Tokens() {
			if token.Kind == govaluate.MODIFIER {
				return fmt.Errorf("forbidden modifier tokens in statement '%s'", statement)
			}
		}
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
