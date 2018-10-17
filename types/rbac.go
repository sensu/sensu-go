package types

import (
	"errors"
	"fmt"
	"net/url"
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

	// RuleTypeAsset access control for asset objects
	RuleTypeAsset = "assets"

	// RuleTypeCheck access control for check objects
	RuleTypeCheck = "checks"

	// RuleTypeCluster access control for cluster management
	RuleTypeCluster = "cluster"

	// RuleTypeEntity access control for entity objects
	RuleTypeEntity = "entities"

	// RuleTypeEvent access control for event objects
	RuleTypeEvent = "events"

	// RuleTypeEventFilter access control for filter objects
	RuleTypeEventFilter = "filters"

	// RuleTypeExtension access control for extension registry
	RuleTypeExtension = "extensions"

	// RuleTypeHandler access control for handler objects
	RuleTypeHandler = "handlers"

	// RuleTypeHook access control for hook objects
	RuleTypeHook = "hooks"

	// RuleTypeMutator access control for mutator objects
	RuleTypeMutator = "mutators"

	// RuleTypeNamespace access control for namespace objects
	RuleTypeNamespace = "namespaces"

	// RuleTypeRole access control for role objects
	RuleTypeRole = "roles"

	// RuleTypeSilenced access control for silenced objects
	RuleTypeSilenced = "silenced"

	// RuleTypeUser access control for user objects
	RuleTypeUser = "users"
)

var (
	// RuleAllPerms all actions
	RuleAllPerms = []string{
		RulePermCreate,
		RulePermRead,
		RulePermUpdate,
		RulePermDelete,
	}

	// AllTypes specifies all possible types
	AllTypes = []string{
		RuleTypeAll,
		RuleTypeAsset,
		RuleTypeCheck,
		RuleTypeEntity,
		RuleTypeEvent,
		RuleTypeEventFilter,
		RuleTypeExtension,
		RuleTypeHandler,
		RuleTypeHook,
		RuleTypeMutator,
		RuleTypeNamespace,
		RuleTypeRole,
		RuleTypeSilenced,
		RuleTypeUser,
	}
)

//
// Validators

// Validate returns an error if the rule is invalid.
func (r *Rule) Validate() error {
	if r.Type == "" {
		return errors.New("type can't be empty")
	}

	if r.Namespace != "*" {
		if err := ValidateNameStrict(r.Namespace); err != nil {
			return errors.New("namespace " + err.Error())
		}
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

	for _, rule := range r.Rules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("rule %s", err)
		}

		// TODO: Check for duplicate rule definitions?
	}

	return nil
}

// URIPath returns the path component of a Role URI.
func (r *Role) URIPath() string {
	return fmt.Sprintf("/rbac/roles/%s", url.PathEscape(r.Name))
}

//
// Fixtures

// FixtureRule returns a partial rule
func FixtureRule(namespace string) *Rule {
	return &Rule{
		Type:      RuleTypeAll,
		Namespace: namespace,
		Permissions: []string{
			RulePermCreate,
			RulePermRead,
			RulePermUpdate,
			RulePermDelete,
		},
	}
}

// FixtureRuleWithPerms returns a partial rule with perms applied
func FixtureRuleWithPerms(T string, perms ...string) Rule {
	rule := *FixtureRule("*")
	rule.Type = T
	rule.Permissions = perms
	return rule
}

// FixtureRole returns a partial role
func FixtureRole(name, namespace string) *Role {
	return &Role{
		Name: name,
		Rules: []Rule{
			*FixtureRule(namespace),
		},
	}
}
