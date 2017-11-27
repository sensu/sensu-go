package etcd

import (
	"context"
	"path"

	"github.com/sensu/sensu-go/types"
)

const keySeparator = "/"

// Easily build multi-tenant resource keys
type keyBuilder struct {
	resourceName string
	namespace    namespace
}

func newKeyBuilder(resourceName string) keyBuilder {
	builder := keyBuilder{resourceName: resourceName}
	return builder
}

func (b keyBuilder) withOrg(org string) keyBuilder {
	ns := namespace{org: org}
	return b.withNamespace(ns)
}

func (b keyBuilder) withResource(r types.MultitenantResource) keyBuilder {
	b.namespace = newNamespaceFromResource(r)
	return b
}

func (b keyBuilder) withContext(ctx context.Context) keyBuilder {
	ns := newNamespaceFromContext(ctx)
	return b.withNamespace(ns)
}

func (b keyBuilder) withNamespace(ns namespace) keyBuilder {
	b.namespace = ns
	return b
}

func (b keybuilder) buildPrefix(keys ...string) string {
	out := etcdRoot + keySeparator + b.resourceName

	keys = append([]string{b.namespace.org, b.namespace.env}, keys...)
	for _, key := range keys {
		// If we encounter a wildcard stop and return key
		if key == wildcardValue {
			return out
		}
		// If the key is nil we ignore and do not add
		if key == "" {
			continue
		}
		out = out + keySeparator + key
	}

	return out
}

func (b keybuilder) build(keys ...string) string {
	items := append([]string{etcdRoot, b.resourceName, b.org, b.env}, keys...)
	return path.Join(items...)
}
