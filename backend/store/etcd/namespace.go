package etcd

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

const wildcardValue = "*"

type namespace struct {
	org string
	env string
}

func newNamespaceFromContext(ctx context.Context) namespace {
	return namespace{
		org: organization(ctx),
		env: environment(ctx),
	}
}

func newNamespaceFromResource(resource types.MultitenantResource) namespace {
	return namespace{
		org: resource.GetOrganization(),
		env: resource.GetEnvironment(),
	}
}

func (ns namespace) OrgIsWildcard() bool {
	return ns.org == wildcardValue
}

func (ns namespace) EnvIsWildcard() bool {
	return ns.env == wildcardValue
}

func (ns namespace) Wildcard() bool {
	return ns.EnvIsWildcard() && ns.OrgIsWildcard()
}
