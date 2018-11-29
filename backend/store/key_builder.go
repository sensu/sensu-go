package store

import (
	"context"
	"path"

	"github.com/sensu/sensu-go/types"
)

const keySeparator = "/"

// KeyBuilder builds multi-tenant resource keys.
type KeyBuilder struct {
	resourceName string
	namespace    string
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
