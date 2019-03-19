package store

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// PageSizeFromContext returns the page size stored in the given context, if
// any. Returns 0 if none is found, typically meaning "unlimited" page size for
// the store.
func PageSizeFromContext(ctx context.Context) int {
	if value := ctx.Value(corev2.PageSizeKey); value != nil {
		return value.(int)
	}
	return 0
}

// PageContinueFromContext returns the continue token stored in the given
// context, if any. Returns "" if none is found.
func PageContinueFromContext(ctx context.Context) string {
	if value := ctx.Value(corev2.PageContinueKey); value != nil {
		return value.(string)
	}
	return ""
}
