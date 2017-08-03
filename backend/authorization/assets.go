package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Assets is global instance of AssetPolicy
var Assets = AssetPolicy{}

// AssetPolicy ...
type AssetPolicy struct {
	context Context
}

// Resource this policy is associated with
func (u *AssetPolicy) Resource() string {
	return types.RuleTypeAsset
}

// Context(ual) info this instance of the policy is associated with
func (u *AssetPolicy) Context() Context {
	return u.context
}

// WithContext returns new policy populated with rules & organization.
func (p AssetPolicy) WithContext(ctx context.Context) AssetPolicy {
	p.context = ExtractValueFromContext(ctx)
	return p
}

// CanList returns true if actor has read access to resource.
func (p *AssetPolicy) CanList() bool {
	return canPerform(p, types.RulePermRead)
}

// CanRead returns true if actor has read access to resource.
func (p *AssetPolicy) CanRead(asset *types.Asset) bool {
	return canPerformOn(p, asset.Organization, types.RulePermRead)
}

// CanCreate returns true if actor has access to create.
func (p *AssetPolicy) CanCreate() bool {
	return canPerform(p, types.RulePermCreate)
}

// CanUpdate returns true if actor has access to update.
func (p *AssetPolicy) CanUpdate() bool {
	return canPerform(p, types.RulePermUpdate)
}

// CanDelete returns true if actor has access to delete.
func (p *AssetPolicy) CanDelete() bool {
	return canPerform(p, types.RulePermDelete)
}
