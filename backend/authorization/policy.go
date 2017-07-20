package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Actor describes an entity who can perform actions within the system that are
// bound by access controls.
type Actor struct {
	Organization string
	Roles        []*types.Role
}

// NewActorFromContext given request context returns new actor.
func NewActorFromContext(ctx context.Context) Actor {
	actor := Actor{}

	if organization, ok := ctx.Value(types.OrganizationKey).(string); ok {
		actor.Organization = organization
	}

	if roles, ok := ctx.Value(ContextRoleKey).([]*types.Role); ok {
		actor.Roles = roles
	}

	return actor
}

// Ability encapsulates the abilities a user can perform on a resource.
type Ability struct {
	Resource string
	Actor
}

// WithContext returns new Ability populated with rules & organization.
func (ability Ability) WithContext(ctx context.Context) Ability {
	ability.Actor = NewActorFromContext(ctx)
	return ability
}

// CanRead returns true if actor has read access to resource.
func (abilityPtr *Ability) CanRead() bool {
	return abilityPtr.canPerform(types.RulePermRead)
}

// CanCreate returns true if actor has create access to resource.
func (abilityPtr *Ability) CanCreate() bool {
	return abilityPtr.canPerform(types.RulePermCreate)
}

// CanUpdate returns true if actor has update access to resource.
func (abilityPtr *Ability) CanUpdate() bool {
	return abilityPtr.canPerform(types.RulePermUpdate)
}

// CanDelete returns true if actor has update access to resource.
func (abilityPtr *Ability) CanDelete() bool {
	return abilityPtr.canPerform(types.RulePermDelete)
}

func (abilityPtr *Ability) canPerform(action string) bool {
	return CanAccessResource(
		abilityPtr.Actor,
		abilityPtr.Resource,
		action,
	)
}
