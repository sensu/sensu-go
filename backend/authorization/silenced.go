package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Silenced is global instance of SilencedPolicy
var Silenced = SilencedPolicy{}

// SilencedPolicy ...
type SilencedPolicy struct {
	context Context
}

// Resource this policy is associated with
func (p *SilencedPolicy) Resource() string {
	return types.RuleTypeSilenced
}

// Context info this instance of the policy is associated with
func (p *SilencedPolicy) Context() Context {
	return p.context
}

// WithContext returns new policy populated with rules & namespace.
func (p SilencedPolicy) WithContext(ctx context.Context) SilencedPolicy { // nolint
	p.context = ExtractValueFromContext(ctx)
	return p
}

// CanList returns true if actor has read access to resource.
func (p *SilencedPolicy) CanList() bool {
	return canPerform(p, types.RulePermRead)
}

// CanRead returns true if actor has read access to resource.
func (p *SilencedPolicy) CanRead(silenced *types.Silenced) bool {
	return canPerformOn(p, silenced.Namespace, types.RulePermRead)
}

// CanCreate returns true if actor has access to create.
func (p *SilencedPolicy) CanCreate(silenced *types.Silenced) bool {
	return canPerformOn(p, silenced.Namespace, types.RulePermCreate)
}

// CanUpdate returns true if actor has access to update.
func (p *SilencedPolicy) CanUpdate(silenced *types.Silenced) bool {
	return canPerformOn(p, silenced.Namespace, types.RulePermUpdate)
}

// CanDelete returns true if actor has access to delete.
func (p *SilencedPolicy) CanDelete() bool {
	return canPerform(p, types.RulePermDelete)
}
