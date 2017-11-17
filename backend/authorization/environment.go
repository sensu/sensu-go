package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Environments is global instance of EnvironmentPolicy
var Environments = EnvironmentPolicy{}

// EnvironmentPolicy ...
type EnvironmentPolicy struct {
	context Context
}

// Resource this policy is associated with
func (p *EnvironmentPolicy) Resource() string {
	return types.RuleTypeEnvironment
}

// Context info this instance of the policy is associated with
func (p *EnvironmentPolicy) Context() Context {
	return p.context
}

// WithContext returns new policy populated with rules & organization.
func (p EnvironmentPolicy) WithContext(ctx context.Context) EnvironmentPolicy { // nolint
	p.context = ExtractValueFromContext(ctx)
	return p
}

// CanList returns true if actor has read access to resource.
func (p *EnvironmentPolicy) CanList() bool {
	return canPerform(p, types.RulePermRead)
}

// CanRead returns true if actor has read access to resource.
func (p *EnvironmentPolicy) CanRead(env *types.Environment) bool {
	return canPerformOn(p, env.Organization, env.Name, types.RulePermRead)
}

// CanCreate returns true if actor has access to create.
func (p *EnvironmentPolicy) CanCreate(env *types.Environment) bool {
	return canPerformOn(p, env.Organization, env.Name, types.RulePermCreate)
}

// CanUpdate returns true if actor has access to update.
func (p *EnvironmentPolicy) CanUpdate(env *types.Environment) bool {
	return canPerformOn(p, env.Organization, env.Name, types.RulePermUpdate)
}

// CanDelete returns true if actor has access to delete.
func (p *EnvironmentPolicy) CanDelete() bool {
	return canPerform(p, types.RulePermDelete)
}
