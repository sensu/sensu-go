package store

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

const (
	// WildcardValue is the symbol that denotes a wildcard organization or environment.
	WildcardValue = "*"

	// Root is the root of the sensu keyspace.
	Root = "/sensu.io"
)

// Namespace describes the values in-which a Sensu resource may reside.
type Namespace struct {
	Org string
	Env string
}

// NewNamespaceFromContext creates a new Namespace from a context.
func NewNamespaceFromContext(ctx context.Context) Namespace {
	return Namespace{
		Org: organization(ctx),
		Env: environment(ctx),
	}
}

// NewNamespaceFromResource creates a new Namespace from a MultitenantResource.
func NewNamespaceFromResource(resource types.MultitenantResource) Namespace {
	return Namespace{
		Org: resource.GetOrganization(),
		Env: resource.GetEnvironment(),
	}
}

// OrgIsWildcard returns true if the organization is a wildcard.
func (ns Namespace) OrgIsWildcard() bool {
	return ns.Org == WildcardValue
}

// EnvIsWildcard returns true if the environment is a wildcard.
func (ns Namespace) EnvIsWildcard() bool {
	return ns.Env == WildcardValue
}

// Wildcard returns true if all of the namespace values are wildcards.
func (ns Namespace) Wildcard() bool {
	return ns.EnvIsWildcard() && ns.OrgIsWildcard()
}

// environment returns the environment name injected in the context
func environment(ctx context.Context) string {
	if value := ctx.Value(types.EnvironmentKey); value != nil {
		return value.(string)
	}
	return ""
}

// organization returns the organization name injected in the context
func organization(ctx context.Context) string {
	if value := ctx.Value(types.OrganizationKey); value != nil {
		return value.(string)
	}
	return ""
}
