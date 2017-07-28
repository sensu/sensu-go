package middlewares

import "net/http"

// HTTPMiddleware interface serves as the building block for composing
// http middleware.
type HTTPMiddleware interface {
	// Then returns new handler that wraps given handler
	Then(http.Handler) http.Handler
}
