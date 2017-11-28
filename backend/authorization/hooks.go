package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Hooks is global instance of HookPolicy
var Hooks = HookPolicy{}

// HookPolicy ...
type HookPolicy struct {
	context Context
}

// Resource this policy is associated with
func (p *HookPolicy) Resource() string {
	return types.RuleTypeHook
}

// Context info this instance of the policy is associated with
func (p *HookPolicy) Context() Context {
	return p.context
}

// WithContext returns new policy populated with rules & organization.
func (p HookPolicy) WithContext(ctx context.Context) HookPolicy { // nolint
	p.context = ExtractValueFromContext(ctx)
	return p
}

// CanList returns true if actor has read access to resource.
func (p *HookPolicy) CanList() bool {
	return canPerform(p, types.RulePermRead)
}

// CanRead returns true if actor has read access to resource.
func (p *HookPolicy) CanRead(hook *types.HookConfig) bool {
	return canPerformOn(p, hook.Organization, hook.Environment, types.RulePermRead)
}

// CanCreate returns true if actor has access to create.
func (p *HookPolicy) CanCreate(hook *types.HookConfig) bool {
	return canPerformOn(p, hook.Organization, hook.Environment, types.RulePermCreate)
}

// CanUpdate returns true if actor has access to update.
func (p *HookPolicy) CanUpdate(hook *types.HookConfig) bool {
	return canPerformOn(p, hook.Organization, hook.Environment, types.RulePermUpdate)
}

// CanDelete returns true if actor has access to delete.
func (p *HookPolicy) CanDelete() bool {
	return canPerform(p, types.RulePermDelete)
}
