package etcd

import (
	"context"
	"path"

	"github.com/sensu/sensu-go/types"
)

// Easily build multi-tenant resource keys
type keyBuilder struct {
	org          string
	env          string
	resourceName string
}

func newKeyBuilder(resourceName string) keyBuilder {
	builder := keyBuilder{resourceName: resourceName}
	return builder
}

func (b keyBuilder) withOrg(org string) keyBuilder {
	b.org = org
	return b
}

func (b keyBuilder) withResource(r types.MultitenantResource) keyBuilder {
	b.org = r.GetOrg()
	b.env = r.GetEnv()
	return b
}

func (b keyBuilder) withContext(ctx context.Context) keyBuilder {
	b.env = environment(ctx)
	b.org = organization(ctx)
	return b
}

func (b keyBuilder) build(keys ...string) string {
	items := append([]string{etcdRoot, b.resourceName, b.org, b.env}, keys...)
	return path.Join(items...)
}
