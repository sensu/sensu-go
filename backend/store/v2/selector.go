package v2

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/selector"
)

// EventContextWithSelector returns a new context, with the selector stored as a
// value.
func EventContextWithSelector(ctx context.Context, selector *selector.Selector) context.Context {
	return ContextWithSelector(ctx, corev2.TypeMeta{APIVersion: "core/v2", Type: "Event"}, selector)
}

// EventSelectorFromContext extracts the selector stored as a context value, if it
// exists.
func EventSelectorFromContext(ctx context.Context) *selector.Selector {
	return SelectorFromContext(ctx, corev2.TypeMeta{APIVersion: "core/v2", Type: "Event"})
}

type selectorCtxKey struct {
	Type       string
	APIVersion string
}

// ContextWithSelector returns a new context, with the selector stored
// as a value for a specific resource type.
func ContextWithSelector(ctx context.Context, tm corev2.TypeMeta, selector *selector.Selector) context.Context {
	return context.WithValue(ctx, selectorCtxKey{Type: tm.Type, APIVersion: tm.APIVersion}, selector)
}

// SelectorFromContext extracts the selector stored in context for a
// specific resource type, if it exists.
func SelectorFromContext(ctx context.Context, tm corev2.TypeMeta) *selector.Selector {
	val := ctx.Value(selectorCtxKey{Type: tm.Type, APIVersion: tm.APIVersion})
	if val == nil {
		return nil
	}
	return val.(*selector.Selector)
}
