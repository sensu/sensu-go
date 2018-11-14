package types

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/sensu/sensu-go/js"
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

// FixtureEventFilter returns a Filter fixture for testing.
func FixtureEventFilter(name string) *EventFilter {
	return &EventFilter{
		Action:      EventFilterActionAllow,
		Expressions: []string{"event.check.team == 'ops'"},
		ObjectMeta: ObjectMeta{
			Namespace: "default",
			Name:      name,
		},
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

// URIPath returns the path component of a Filter URI.
func (f *EventFilter) URIPath() string {
	return fmt.Sprintf("/filters/%s", url.PathEscape(f.Name))
}
