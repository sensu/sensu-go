package store

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

const (
	// WildcardValue is the symbol that denotes a wildcard namespace.
	WildcardValue = "*"

	// Root is the root of the sensu keyspace.
	Root = "/sensu.io"
)

// Namespace describes the values in-which a Sensu resource may reside.
type Namespace struct {
	Namespace string
}

// NewNamespaceFromContext creates a new Namespace from a context.
func NewNamespaceFromContext(ctx context.Context) Namespace {
	return Namespace{
		Namespace: namespace(ctx),
	}
}

// NewNamespaceFromResource creates a new Namespace from a MultitenantResource.
func NewNamespaceFromResource(resource types.MultitenantResource) Namespace {
	return Namespace{
		Namespace: resource.GetNamespace(),
	}
}

// NamespaceIsWildcard returns true if the namespace is a wildcard.
func (ns Namespace) NamespaceIsWildcard() bool {
	return ns.Namespace == WildcardValue
}

// namespace returns the namespace name injected in the context
func namespace(ctx context.Context) string {
	if value := ctx.Value(types.NamespaceKey); value != nil {
		return value.(string)
	}
	return ""
}
