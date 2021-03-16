package v2

import "context"

// MultitenantResource is a object that belongs to a namespace
type MultitenantResource interface {
	GetNamespace() string
}

// SetContextFromResource takes a context and a multi-tenant resource, adds the
// namespace to the context, and returns the udpated context
func SetContextFromResource(ctx context.Context, r MultitenantResource) context.Context {
	ctx = context.WithValue(ctx, NamespaceKey, r.GetNamespace())
	return ctx
}
