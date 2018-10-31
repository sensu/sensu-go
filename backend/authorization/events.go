package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Events is global instance of EventPolicy
var Events = EventPolicy{}

// EventPolicy ...
type EventPolicy struct {
	context Context
}

// Resource this policy is associated with
func (p *EventPolicy) Resource() string {
	return types.RuleTypeEvent
}

// Context info this instance of the policy is associated with
func (p *EventPolicy) Context() Context {
	return p.context
}

// WithContext returns new policy populated with rules & namespace.
func (p EventPolicy) WithContext(ctx context.Context) EventPolicy { // nolint
	p.context = ExtractValueFromContext(ctx)
	return p
}

// CanList returns true if actor has read access to resource.
func (p *EventPolicy) CanList() bool {
	return canPerform(p, types.RulePermRead)
}

// CanRead returns true if actor has read access to resource.
func (p *EventPolicy) CanRead(event *types.Event) bool {
	return canPerformOn(p, event.Entity.Namespace, types.RulePermRead)
}

// CanCreate returns true if actor has access to create.
func (p *EventPolicy) CanCreate(event *types.Event) bool {
	return canPerformOn(p, event.Entity.Namespace, types.RulePermCreate)
}

// CanUpdate returns true if actor has access to update.
func (p *EventPolicy) CanUpdate(event *types.Event) bool {
	return canPerformOn(p, event.Entity.Namespace, types.RulePermUpdate)
}

// CanDelete returns true if actor has access to delete.
func (p *EventPolicy) CanDelete() bool {
	return canPerform(p, types.RulePermDelete)
}
