package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Organizations is global instance of OrganizationPolicy
var Organizations = OrganizationPolicy{}

// OrganizationPolicy ...
type OrganizationPolicy struct {
	context Context
}

// Resource this policy is associated with
func (p *OrganizationPolicy) Resource() string {
	return types.RuleTypeOrganization
}

// Context info this instance of the policy is associated with
func (p *OrganizationPolicy) Context() Context {
	return p.context
}

// WithContext returns new policy populated with rules & organization.
func (p OrganizationPolicy) WithContext(ctx context.Context) OrganizationPolicy { // nolint
	p.context = ExtractValueFromContext(ctx)
	p.context.Organization = "*"

	return p
}

// CanList returns true if actor has read access to resource.
func (p *OrganizationPolicy) CanList() bool {
	return canPerform(p, types.RulePermRead)
}

// CanRead returns true if actor has read access to resource.
func (p *OrganizationPolicy) CanRead(org *types.Organization) bool {
	return CanAccessResource(
		p.Context().Actor,
		org.Name,
		"",
		p.Resource(),
		types.RulePermRead,
	)
}

// CanCreate returns true if actor has access to create.
func (p *OrganizationPolicy) CanCreate(org *types.Organization) bool {
	return canPerform(p, types.RulePermCreate)
}

// CanUpdate returns true if actor has access to update.
func (p *OrganizationPolicy) CanUpdate(org *types.Organization) bool {
	return canPerform(p, types.RulePermUpdate)
}

// CanDelete returns true if actor has access to delete.
func (p *OrganizationPolicy) CanDelete() bool {
	return canPerform(p, types.RulePermDelete)
}
