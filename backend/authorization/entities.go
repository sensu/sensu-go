package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Entities is global instance of EntityPolicy
var Entities = EntityPolicy{}

// EntityPolicy ...
type EntityPolicy struct {
	context Context
}

// Resource this policy is associated with
func (p *EntityPolicy) Resource() string {
	return types.RuleTypeEntity
}

// Context info this instance of the policy is associated with
func (p *EntityPolicy) Context() Context {
	return p.context
}

// WithContext returns new policy populated with rules & namespace.
func (p EntityPolicy) WithContext(ctx context.Context) EntityPolicy { // nolint
	p.context = ExtractValueFromContext(ctx)
	return p
}

// CanList returns true if actor has read access to resource.
func (p *EntityPolicy) CanList() bool {
	return canPerform(p, types.RulePermRead)
}

// CanRead returns true if actor has read access to resource.
func (p *EntityPolicy) CanRead(entity *types.Entity) bool {
	return canPerformOn(p, entity.Namespace, types.RulePermRead)
}

// CanCreate returns true if actor has access to create.
func (p *EntityPolicy) CanCreate(entity *types.Entity) bool {
	return canPerformOn(p, entity.Namespace, types.RulePermCreate)
}

// CanUpdate returns true if actor has access to update.
func (p *EntityPolicy) CanUpdate(entity *types.Entity) bool {
	return canPerformOn(p, entity.Namespace, types.RulePermUpdate)
}

// CanDelete returns true if actor has access to delete.
func (p *EntityPolicy) CanDelete() bool {
	return canPerform(p, types.RulePermDelete)
}
