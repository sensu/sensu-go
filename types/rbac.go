package types

import (
	"errors"
	"fmt"
)

const (
	// RuleTypeAll matches all actions
	RuleTypeAll = "*"

	// RulePermCreate create action
	RulePermCreate = "create"

	// RulePermRead read action
	RulePermRead = "read"

	// RulePermUpdate update action
	RulePermUpdate = "update"

	// RulePermDelete delete action
	RulePermDelete = "delete"
)

// Rule maps permissions to a given type
type Rule struct {
	Type        string   `json:"type"`
	Permissions []string `json:"permissions"`
}

// Role describes set of rules
type Role struct {
	Name         string `json:"name"`
	Organization string `json:"organization"`
	Rules        []Rule `json:"rules"`
}

//
// Validators

// Validate returns an error if the rule is invalid.
func (r *Rule) Validate() error {
	if r.Type == "" {
		return errors.New("type can't be empty")
	}

	if len(r.Permissions) == 0 {
		return errors.New("permissions must have at least one permission")
	}

	for _, p := range r.Permissions {
		switch p {
		case RulePermCreate, RulePermRead, RulePermUpdate, RulePermDelete:
		default:
			return fmt.Errorf(
				"permission '%s' is not valid - must be one of ['%s', '%s', '%s', '%s']",
				p,
				RulePermCreate,
				RulePermRead,
				RulePermUpdate,
				RulePermDelete,
			)
		}
	}

	return nil
}

// Validate returns an error if the role is invalid.
func (r *Role) Validate() error {
	if err := ValidateNameStrict(r.Name); err != nil {
		return errors.New("name " + err.Error())
	}

	if err := ValidateNameStrict(r.Organization); err != nil {
		return errors.New("organization " + err.Error())
	}

	if len(r.Rules) == 0 {
		return errors.New("rules must at least contain one element")
	}

	return nil
}

//
// Fixtures

// FixtureRule returns a partial rule
func FixtureRule() *Rule {
	return &Rule{
		Type: RuleTypeAll,
		Permissions: []string{
			RulePermCreate,
			RulePermRead,
			RulePermUpdate,
			RulePermDelete,
		},
	}
}

// FixtureRole returns a partial role
func FixtureRole(name string, org string) *Role {
	return &Role{
		Name:         name,
		Organization: org,
		Rules: []Rule{
			*FixtureRule(),
		},
	}
}
