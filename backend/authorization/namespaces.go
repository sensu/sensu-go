package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Namespaces is global instance of NamespacePolicy
var Namespaces = NamespacePolicy{}

// NamespacePolicy ...
type NamespacePolicy struct {
	context Context
}

// Resource this policy is associated with
func (p *NamespacePolicy) Resource() string {
	return types.RuleTypeNamespace
}

// Context info this instance of the policy is associated with
func (p *NamespacePolicy) Context() Context {
	return p.context
}

// WithContext returns new policy populated with rules & namespace.
func (p NamespacePolicy) WithContext(ctx context.Context) NamespacePolicy { // nolint
	p.context = ExtractValueFromContext(ctx)
	p.context.Namespace = types.NamespaceTypeAll

	return p
}

// CanList returns true if actor has read access to resource.
func (p *NamespacePolicy) CanList() bool {
	return canPerform(p, types.RulePermRead)
}

// CanRead returns true if actor has read access to resource.
func (p *NamespacePolicy) CanRead(namespace *types.Namespace) bool {
	return CanAccessResource(
		p.Context().Actor,
		namespace.Name,
		p.Resource(),
		types.RulePermRead,
	)
}

// CanCreate returns true if actor has access to create.
func (p *NamespacePolicy) CanCreate(namespace *types.Namespace) bool {
	return canPerform(p, types.RulePermCreate)
}

// CanUpdate returns true if actor has access to update.
func (p *NamespacePolicy) CanUpdate(namespace *types.Namespace) bool {
	return canPerform(p, types.RulePermUpdate)
}

// CanDelete returns true if actor has access to delete.
func (p *NamespacePolicy) CanDelete() bool {
	return canPerform(p, types.RulePermDelete)
}
