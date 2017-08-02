package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Actor describes an entity who can perform actions within the system that are
// bound by access controls.
type Actor struct {
	Name  string
	Rules []types.Rule
}

// Context holds the organization the action is associated with and the user
// making said action.
type Context struct {
	Actor        Actor
	Organization string
}

// ExtractAuthoriationContext extracts authorization details from a context
func ExtractValueFromContext(ctx context.Context) Context {
	context := Context{}

	if organization, ok := ctx.Value(types.OrganizationKey).(string); ok {
		context.Organization = organization
	}

	if actor, ok := ctx.Value(types.AuthorizationActorKey).(Actor); ok {
		context.Actor = actor
	}

	return context
}

// Ability encapsulates the abilities a user can perform on a resource.
type Ability struct {
	Resource     string
	Organization string
	Actor
}

// WithContext returns new Ability populated with rules & organization.
func (ability Ability) WithContext(ctx context.Context) Ability {
	v := ExtractValueFromContext(ctx)
	ability.Actor = v.Actor // TODO: RIP
	ability.Organization = v.Organization
	return ability
}

// CanRead returns true if actor has read access to resource.
func (abilityPtr *Ability) CanRead() bool { // nolint
	return abilityPtr.canPerform(types.RulePermRead)
}

// CanCreate returns true if actor has create access to resource.
func (abilityPtr *Ability) CanCreate() bool { // nolint
	return abilityPtr.canPerform(types.RulePermCreate)
}

// CanUpdate returns true if actor has update access to resource.
func (abilityPtr *Ability) CanUpdate() bool { // nolint
	return abilityPtr.canPerform(types.RulePermUpdate)
}

// CanDelete returns true if actor has update access to resource.
func (abilityPtr *Ability) CanDelete() bool { // nolint
	return abilityPtr.canPerform(types.RulePermDelete)
}

func (abilityPtr *Ability) canPerform(action string) bool { // nolint
	return CanAccessResource(
		abilityPtr.Actor,
		abilityPtr.Organization,
		abilityPtr.Resource,
		action,
	)
}

// Policy ...
type Policy interface { // TODO: rename to ...?
	Resource() string
	Context() Context
}

func canPerform(policy Policy, action string) bool {
	return CanAccessResource(
		policy.Context().Actor,
		policy.Context().Organization,
		policy.Resource(),
		action,
	)
}

func canPerformOn(policy Policy, org, action string) bool {
	if policy.Context().Organization != org {
		return false
	}

	return canPerform(policy, action)
}
