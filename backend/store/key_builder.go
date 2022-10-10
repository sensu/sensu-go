package store

import (
	"context"
	"path"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/types"
)

const keySeparator = "/"

// KeyBuilder builds multi-tenant resource keys.
type KeyBuilder struct {
	resourceName         string
	namespace            string
	includeTrailingSlash bool
}

// NewKeyBuilder creates a new KeyBuilder.
func NewKeyBuilder(resourceName string) KeyBuilder {
	builder := KeyBuilder{resourceName: resourceName}
	return builder
}

// WithNamespace adds a namespace to a key.
func (b KeyBuilder) WithNamespace(namespace string) KeyBuilder {
	b.namespace = namespace
	return b
}

// WithResource adds a resource to a key.
func (b KeyBuilder) WithResource(r types.MultitenantResource) KeyBuilder {
	b.namespace = NewNamespaceFromResource(r)
	return b
}

// WithContext adds a namespace from a context.
func (b KeyBuilder) WithContext(ctx context.Context) KeyBuilder {
	b.namespace = NewNamespaceFromContext(ctx)
	return b
}

// WithExactMatch ensures the final key has a trailing slash.
// This is useful for searching sub-resources without erroneously including
// resources that the final path component is a prefix for.
// eg consider two entities, foo and foobar, both with events.
// Without a trailing slash, we end up iterating over both
// /foo/event and /foobar/event.
func (b KeyBuilder) WithExactMatch() KeyBuilder {
	b.includeTrailingSlash = true
	return b
}

// Build builds a key from the components it is given.
func (b KeyBuilder) Build(keys ...string) string {
	items := append(
		[]string{
			Root,
			b.resourceName,
			b.namespace,
		},
		keys...,
	)

	key := path.Join(items...)

	// In order to not inadvertently build a key that could list across
	// namespaces, we need to make sure that we terminate the key with the key
	// separator when a namespace is involved without a specific object name
	// within it.
	if b.namespace != "" {
		if len(keys) == 0 || keys[len(keys)-1] == "" {
			key += keySeparator
		}
	}

	// Be specific when listing sub-resources for a specific resource.
	if b.includeTrailingSlash {
		key += keySeparator
	}

	return key
}

// BuildPrefix can be used when building a key that will be used to retrieve a collection of
// records. Unlike standard build method it stops building the key when it first
// encounters a wildcard value.
func (b KeyBuilder) BuildPrefix(keys ...string) string {
	out := Root + keySeparator + b.resourceName

	keys = append([]string{b.namespace}, keys...)
	for _, key := range keys {
		// If we encounter a wildcard stop and return key
		if key == WildcardValue {
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

// KeyFromResource determines the path to a resource in the store using the
// resource itself
func KeyFromResource(r corev2.Resource) string {
	resourcePrefix := r.StorePrefix()
	namespace := r.GetObjectMeta().Namespace
	name := r.GetObjectMeta().Name
	key := path.Join(Root, resourcePrefix, namespace, name)

	// In order to not inadvertently build a key that could list across
	// namespaces, we need to make sure that we terminate the key with the key
	// separator when a namespace is involved without a specific object name
	// within it.
	if namespace != "" {
		if len(name) == 0 {
			key += keySeparator
		}
	}

	return key
}

// KeyFromArgs determines the path to a resource in the store using the provided
// arguments
func KeyFromArgs(ctx context.Context, resourcePrefix, name string) string {
	namespace := NewNamespaceFromContext(ctx)
	key := path.Join(Root, resourcePrefix, namespace, name)

	// In order to not inadvertently build a key that could list across
	// namespaces, we need to make sure that we terminate the key with the key
	// separator when a namespace is involved without a specific object name
	// within it.
	if namespace != "" {
		if len(name) == 0 {
			key += keySeparator
		}
	}

	return key
}
