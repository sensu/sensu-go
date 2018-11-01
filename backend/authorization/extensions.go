package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Extensions is global instance of ExtensionPolicy
var Extensions = ExtensionPolicy{}

// ExtensionPolicy ...
type ExtensionPolicy struct {
	context Context
}

// Resource this policy is associated with
func (p *ExtensionPolicy) Resource() string {
	return types.RuleTypeExtension
}

// Context info this instance of the policy is associated with
func (p *ExtensionPolicy) Context() Context {
	return p.context
}

// WithContext returns new policy populated with rules & namespace.
func (p ExtensionPolicy) WithContext(ctx context.Context) ExtensionPolicy { // nolint
	p.context = ExtractValueFromContext(ctx)
	return p
}

// CanList returns true if actor has read access to resource.
func (p *ExtensionPolicy) CanList() bool {
	return canPerform(p, types.RulePermRead)
}

// CanRead returns true if actor has read access to resource.
func (p *ExtensionPolicy) CanRead(extension *types.Extension) bool {
	return canPerformOn(p, extension.Namespace, types.RulePermRead)
}

// CanCreate returns true if actor has access to create.
func (p *ExtensionPolicy) CanCreate() bool {
	return canPerform(p, types.RulePermCreate)
}

// CanUpdate returns true if actor has access to update.
func (p *ExtensionPolicy) CanUpdate() bool {
	return canPerform(p, types.RulePermUpdate)
}

// CanDelete returns true if actor has access to delete.
func (p *ExtensionPolicy) CanDelete() bool {
	return canPerform(p, types.RulePermDelete)
}
