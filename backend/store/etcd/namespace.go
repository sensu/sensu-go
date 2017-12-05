package etcd

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

const wildcardValue = "*"

// namespace describes the values in-which a Sensu resource may reside.
type namespace struct {
	org string
	env string
}

// newNamespaceFromContext returns new context given context.
func newNamespaceFromContext(ctx context.Context) namespace {
	return namespace{
		org: organization(ctx),
		env: environment(ctx),
	}
}

// newNamespaceFromContext returns new context given multi-tenant resource.
func newNamespaceFromResource(resource types.MultitenantResource) namespace {
	return namespace{
		org: resource.GetOrganization(),
		env: resource.GetEnvironment(),
	}
}

// OrgIsWildcard returns true if the organization is a wildcard.
func (ns namespace) OrgIsWildcard() bool {
	return ns.org == wildcardValue
}

// EnvIsWildcard returns true if the environment is a wildcard.
func (ns namespace) EnvIsWildcard() bool {
	return ns.env == wildcardValue
}

// Wildcard returns true if all of the namespace values are wildcards.
func (ns namespace) Wildcard() bool {
	return ns.EnvIsWildcard() && ns.OrgIsWildcard()
}
