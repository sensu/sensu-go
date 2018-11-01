package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Filters is global instance of FilterPolicy
var Filters = FilterPolicy{}

// FilterPolicy ...
type FilterPolicy struct {
	context Context
}

// Resource this policy is associated with
func (p *FilterPolicy) Resource() string {
	return types.RuleTypeEventFilter
}

// Context info this instance of the policy is associated with
func (p *FilterPolicy) Context() Context {
	return p.context
}

// WithContext returns new policy populated with rules & namespace.
func (p FilterPolicy) WithContext(ctx context.Context) FilterPolicy { // nolint
	p.context = ExtractValueFromContext(ctx)
	return p
}

// CanList returns true if actor has read access to resource.
func (p *FilterPolicy) CanList() bool {
	return canPerform(p, types.RulePermRead)
}

// CanRead returns true if actor has read access to resource.
func (p *FilterPolicy) CanRead(filter *types.EventFilter) bool {
	return canPerformOn(p, filter.Namespace, types.RulePermRead)
}

// CanCreate returns true if actor has access to create.
func (p *FilterPolicy) CanCreate(filter *types.EventFilter) bool {
	return canPerformOn(p, filter.Namespace, types.RulePermCreate)
}

// CanUpdate returns true if actor has access to update.
func (p *FilterPolicy) CanUpdate(filter *types.EventFilter) bool {
	return canPerformOn(p, filter.Namespace, types.RulePermUpdate)
}

// CanDelete returns true if actor has access to delete.
func (p *FilterPolicy) CanDelete() bool {
	return canPerform(p, types.RulePermDelete)
}
