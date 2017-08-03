package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Roles is global instance of RolePolicy
var Roles = RolePolicy{}

// RolePolicy ...
type RolePolicy struct {
	context Context
}

// Resource this policy is associated with
func (u *RolePolicy) Resource() string {
	return types.RuleTypeRole
}

// Context(ual) info this instance of the policy is associated with
func (u *RolePolicy) Context() Context {
	return u.context
}

// WithContext returns new policy populated with rules & organization.
func (p RolePolicy) WithContext(ctx context.Context) RolePolicy {
	p.context = ExtractValueFromContext(ctx)
	p.context.Organization = "*"

	return p
}

// CanList returns true if actor has read access to resource.
func (p *RolePolicy) CanList() bool {
	return true
}

// CanRead returns true if actor has read access to resource.
func (p *RolePolicy) CanRead(r *types.Role) bool {
	if canPerform(p, types.RulePermRead) {
		return true
	}

	// TODO: May want to allow users to view roles associated w/ their account.
	return false
}

// CanCreate returns true if actor has access to create.
func (p *RolePolicy) CanCreate() bool {
	return canPerform(p, types.RulePermCreate)
}

// CanUpdate returns true if actor has access to update.
func (p *RolePolicy) CanUpdate() bool {
	return canPerform(p, types.RulePermUpdate)
}

// CanDelete returns true if actor has access to delete.
func (p *RolePolicy) CanDelete() bool {
	return canPerform(p, types.RulePermDelete)
}
