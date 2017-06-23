package types

import (
	"errors"
	"fmt"
)

type Rule struct {
	Organization string   `json:"organization"`
	Type         string   `json:"type"`
	Permissions  []string `json:"permissions"`
}

const (
	RuleTypeAll = "*"

	RulePermCreate = "create"
	RulePermRead   = "read"
	RulePermUpdate = "update"
	RulePermDelete = "delete"
)

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

// Validate returns an error if the rule is invalid.
func (r *Rule) Validate() error {
	if r.Organization == "" {
		return errors.New("organization can't be empty")
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
