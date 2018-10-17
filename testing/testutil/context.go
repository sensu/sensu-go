package testutil

import (
	"context"

	"github.com/coreos/etcd/store"
	"github.com/sensu/sensu-go/backend/authorization"
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

// ContextWithOrgEnv given org & env returns new contextFn with values added.
func ContextWithNamespace(namespace string) SetContextFn {
	return func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, types.NamespaceKey, namespace)
		return ctx
	}
}

// ContextWithActor instantiates new Actor with given args and returns new
// contextFn w/ actor value applied.
func ContextWithActor(name string, rules ...types.Rule) SetContextFn {
	return func(ctx context.Context) context.Context {
		actor := authorization.Actor{Name: name, Rules: rules}
		return context.WithValue(ctx, types.AuthorizationActorKey, actor)
	}
}

// ContextWithRules instantiates new Actor with given rules and returns new
// contextFn w/ actor value applied.
func ContextWithRules(rules ...types.Rule) SetContextFn {
	return ContextWithActor("fixture", rules...)
}

// ContextWithPerms instantiates new Actor with given rule and returns new
// contextFn w/ actor value applied.
func ContextWithPerms(rule string, perms ...string) SetContextFn {
	return ContextWithRules(types.FixtureRuleWithPerms(rule, perms...))
}

// ContextWithFullAccess instantiates new Actor with full access to resources across
// the system and returns new contextFn w/ actor value applied.
func ContextWithFullAccess(ctx context.Context) context.Context {
	applyContextFn := ContextWithRules(*types.FixtureRule("*"))
	return applyContextFn(ctx)
}

// ContextWithNoAccess instantiates new Actor with no access to resources across
// the system and returns new contextFn w/ actor value applied.
func ContextWithNoAccess(ctx context.Context) context.Context {
	applyContextFn := ContextWithRules()
	return applyContextFn(ctx)
}

// ContextWithStore returns new contextFn with given store value applied to
// context.
func ContextWithStore(store store.Store) SetContextFn {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, types.StoreKey, store)
	}
}
