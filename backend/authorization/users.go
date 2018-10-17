package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Users is global instance of UserPolicy
var Users = UserPolicy{}

// UserPolicy ...
type UserPolicy struct {
	context Context
}

// Resource this policy is associated with
func (p *UserPolicy) Resource() string {
	return types.RuleTypeUser
}

// Context info this instance of the policy is associated with
func (p *UserPolicy) Context() Context {
	return p.context
}

// WithContext returns new policy populated with rules & namespace.
func (p UserPolicy) WithContext(ctx context.Context) UserPolicy { // nolint
	p.context = ExtractValueFromContext(ctx)
	return p
}

// CanList returns true if actor has read access to resource.
func (p *UserPolicy) CanList() bool {
	// Allow users to list but when collection is filtered only allow them to see
	// their own account.
	return true
}

// CanRead returns true if actor has read access to resource.
func (p *UserPolicy) CanRead(user *types.User) bool {
	// Allow users to see their account
	if p.context.Actor.Name == user.Username {
		return true
	}

	return canPerform(p, types.RulePermRead)
}

// CanCreate returns true if actor has access to create.
func (p *UserPolicy) CanCreate() bool {
	return canPerform(p, types.RulePermCreate)
}

// CanUpdate returns true if actor has access to update.
func (p *UserPolicy) CanUpdate(_ *types.User) bool {
	return canPerform(p, types.RulePermUpdate)
}

// CanChangePassword returns true if actor has access to update.
func (p *UserPolicy) CanChangePassword(user *types.User) bool {
	// Allow users to change their password
	if p.context.Actor.Name == user.Username {
		return true
	}

	return canPerform(p, types.RulePermUpdate)
}

// CanDelete returns true if actor has access to delete.
func (p *UserPolicy) CanDelete(user *types.User) bool {
	// Allow users to delete their own account
	if p.context.Actor.Name == user.Username {
		return true
	}

	return canPerform(p, types.RulePermDelete)
}
