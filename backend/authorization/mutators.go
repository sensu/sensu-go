package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Mutators is global instance of MutatorPolicy
var Mutators = MutatorPolicy{}

// MutatorPolicy ...
type MutatorPolicy struct {
	context Context
}

// Resource this policy is associated with
func (p *MutatorPolicy) Resource() string {
	return types.RuleTypeMutator
}

// Context info this instance of the policy is associated with
func (p *MutatorPolicy) Context() Context {
	return p.context
}

// WithContext returns new policy populated with rules & organization.
func (p MutatorPolicy) WithContext(ctx context.Context) MutatorPolicy { // nolint
	p.context = ExtractValueFromContext(ctx)
	return p
}

// CanList returns true if actor has read access to resource.
func (p *MutatorPolicy) CanList() bool {
	return canPerform(p, types.RulePermRead)
}

// CanRead returns true if actor has read access to resource.
func (p *MutatorPolicy) CanRead(mutator *types.Mutator) bool {
	return canPerformOn(p, mutator.Organization, mutator.Environment, types.RulePermRead)
}

// CanCreate returns true if actor has access to create.
func (p *MutatorPolicy) CanCreate() bool {
	return canPerform(p, types.RulePermCreate)
}

// CanUpdate returns true if actor has access to update.
func (p *MutatorPolicy) CanUpdate() bool {
	return canPerform(p, types.RulePermUpdate)
}

// CanDelete returns true if actor has access to delete.
func (p *MutatorPolicy) CanDelete() bool {
	return canPerform(p, types.RulePermDelete)
}
