package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Handlers is global instance of HandlerPolicy
var Handlers = HandlerPolicy{}

// HandlerPolicy ...
type HandlerPolicy struct {
	context Context
}

// Resource this policy is associated with
func (p *HandlerPolicy) Resource() string {
	return types.RuleTypeHandler
}

// Context info this instance of the policy is associated with
func (p *HandlerPolicy) Context() Context {
	return p.context
}

// WithContext returns new policy populated with rules & namespace.
func (p HandlerPolicy) WithContext(ctx context.Context) HandlerPolicy { // nolint
	p.context = ExtractValueFromContext(ctx)
	return p
}

// CanList returns true if actor has read access to resource.
func (p *HandlerPolicy) CanList() bool {
	return canPerform(p, types.RulePermRead)
}

// CanRead returns true if actor has read access to resource.
func (p *HandlerPolicy) CanRead(handler *types.Handler) bool {
	return canPerformOn(p, handler.Namespace, types.RulePermRead)
}

// CanCreate returns true if actor has access to create.
func (p *HandlerPolicy) CanCreate(handler *types.Handler) bool {
	return canPerformOn(p, handler.Namespace, types.RulePermCreate)
}

// CanUpdate returns true if actor has access to update.
func (p *HandlerPolicy) CanUpdate(handler *types.Handler) bool {
	return canPerformOn(p, handler.Namespace, types.RulePermUpdate)
}

// CanDelete returns true if actor has access to delete.
func (p *HandlerPolicy) CanDelete() bool {
	return canPerform(p, types.RulePermDelete)
}
