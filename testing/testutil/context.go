package testutil

import (
	"context"

	"github.com/coreos/etcd/store"
	"github.com/sensu/sensu-go/types"
)

// SetContextFn take context and return new context
type SetContextFn func(context.Context) context.Context

// NewContext instantiates new todo context and applies given contextFns to it.
func NewContext(fns ...SetContextFn) (ctx context.Context) {
	ctx = context.TODO()
	ctx = ApplyContext(ctx, fns...)
	return
}

// ApplyContext applies given contextFns to context.
func ApplyContext(ctx context.Context, fns ...SetContextFn) context.Context {
	for _, fn := range fns {
		ctx = fn(ctx)
	}
	return ctx
}

// ContextWithNamespace returns new contextFn with namespace added.
func ContextWithNamespace(namespace string) SetContextFn {
	return func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, types.NamespaceKey, namespace)
		return ctx
	}
}

// ContextWithStore returns new contextFn with given store value applied to
// context.
func ContextWithStore(store store.Store) SetContextFn {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, types.StoreKey, store)
	}
}
