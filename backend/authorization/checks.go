package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Checks is global instance of CheckPolicy
var Checks = CheckPolicy{}

// CheckPolicy ...
type CheckPolicy struct {
	context Context
}

// Resource this policy is associated with
func (p *CheckPolicy) Resource() string {
	return types.RuleTypeCheck
}

// Context info this instance of the policy is associated with
func (p *CheckPolicy) Context() Context {
	return p.context
}

// WithContext returns new policy populated with rules & namespace.
func (p CheckPolicy) WithContext(ctx context.Context) CheckPolicy { // nolint
	p.context = ExtractValueFromContext(ctx)
	return p
}

// CanList returns true if actor has read access to resource.
func (p *CheckPolicy) CanList() bool {
	return canPerform(p, types.RulePermRead)
}

// CanRead returns true if actor has read access to resource.
func (p *CheckPolicy) CanRead(check *types.CheckConfig) bool {
	return canPerformOn(p, check.Namespace, types.RulePermRead)
}

// CanCreate returns true if actor has access to create.
func (p *CheckPolicy) CanCreate(check *types.CheckConfig) bool {
	return canPerformOn(p, check.Namespace, types.RulePermCreate)
}

// CanUpdate returns true if actor has access to update.
func (p *CheckPolicy) CanUpdate(check *types.CheckConfig) bool {
	return canPerformOn(p, check.Namespace, types.RulePermUpdate)
}

// CanDelete returns true if actor has access to delete.
func (p *CheckPolicy) CanDelete() bool {
	return canPerform(p, types.RulePermDelete)
}
