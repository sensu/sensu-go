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
func (p *RolePolicy) Resource() string {
	return types.RuleTypeRole
}

// Context info this instance of the policy is associated with
func (p *RolePolicy) Context() Context {
	return p.context
}

// WithContext returns new policy populated with rules & organization.
func (p RolePolicy) WithContext(ctx context.Context) RolePolicy { // nolint
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
	// TODO: May want to allow users to view roles associated w/ their account.
	return canPerform(p, types.RulePermRead)
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
