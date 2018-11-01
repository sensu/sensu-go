package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// ClusterPolicy defines access control for cluster administration.
type ClusterPolicy struct {
	context Context
}

// Resource returns the type of resource for this policy
func (p *ClusterPolicy) Resource() string {
	return types.RuleTypeCluster
}

// Context returns the policy context
func (p *ClusterPolicy) Context() Context {
	return p.context
}

// WithContext adds the ctx values to the ClusterPolicy's context
func (p ClusterPolicy) WithContext(ctx context.Context) ClusterPolicy {
	p.context = ExtractValueFromContext(ctx)
	return p
}

// HasPermission returns true if types.RuleAllPerms applies
func (p *ClusterPolicy) HasPermission() bool {
	for _, perm := range types.RuleAllPerms {
		if !canPerform(p, perm) {
			return false
		}
	}
	return true
}
