package selector

import "context"

type selectorContextKey struct{}

// SelectorContextKey is the context key used for passing selectors through
// contexts.
var SelectorContextKey selectorContextKey

// ContextWithSelector returns a new context, with the selector stored as a
// value.
func ContextWithSelector(ctx context.Context, selector *Selector) context.Context {
	return context.WithValue(ctx, SelectorContextKey, selector)
}

// SelectorFromContext extracts the selector stored as a context value, if it
// exists.
func SelectorFromContext(ctx context.Context) *Selector {
	val := ctx.Value(SelectorContextKey)
	if val == nil {
		return nil
	}
	return val.(*Selector)
}
