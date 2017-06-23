package types

import (
	"errors"
	"fmt"
)

const (
	RuleTypeAll = "*"

	RulePermCreate = "create"
	RulePermRead   = "read"
	RulePermUpdate = "update"
	RulePermDelete = "delete"
)

type Rule struct {
	Organization string   `json:"organization"`
	Type         string   `json:"type"`
	Permissions  []string `json:"permissions"`
}

type Role struct {
	Name  string `json:"name"`
	Rules []Rule `json:"rules"`
}

//
// Validators

// Validate returns an error if the rule is invalid.
func (r *Rule) Validate() error {
	if err := ValidateNameStrict(r.Organization); err != nil {
		return errors.New("organization " + err.Error())
	}

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

	if len(r.Rules) == 0 {
		return errors.New("rules must at least contain one element")
	}

	return nil
}

//
// Fixtures

func FixtureRule(organization string) *Rule {
	return &Rule{
		Organization: organization,
		Type:         RuleTypeAll,
		Permissions: []string{
			RulePermCreate,
			RulePermRead,
			RulePermUpdate,
			RulePermDelete,
		},
	}
}

func FixtureRole(name string, org string) *Role {
	return &Role{
		Name: name,
		Rules: []Rule{
			*FixtureRule(org),
		},
	}
}
