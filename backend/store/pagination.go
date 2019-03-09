package store

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// PageSizeFromContext returns the page size stored in the given context, if
// any. Returns 0 if none is found, typically meaning "unlimited" page size for
// the store.
func PageSizeFromContext(ctx context.Context) int {
	if value := ctx.Value(types.PageSizeKey); value != nil {
		return value.(int)
	}
	return 0
}

// PageContinueFromContext returns the continue token stored in the given
// context, if any. Returns "" if none is found.
func PageContinueFromContext(ctx context.Context) string {
	if value := ctx.Value(types.PageContinueKey); value != nil {
		return value.(string)
	}
	return ""
}
