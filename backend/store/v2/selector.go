package v2

import (
	"context"

	"github.com/sensu/sensu-go/backend/selector"
)

type eventSelectorContextKey struct{}

// EventContextWithSelector returns a new context, with the selector stored as a
// value.
func EventContextWithSelector(ctx context.Context, selector *selector.Selector) context.Context {
	return context.WithValue(ctx, eventSelectorContextKey{}, selector)
}

// EventSelectorFromContext extracts the selector stored as a context value, if it
// exists.
func EventSelectorFromContext(ctx context.Context) *selector.Selector {
	val := ctx.Value(eventSelectorContextKey{})
	if val == nil {
		return nil
	}
	return val.(*selector.Selector)
}
