package middlewares

import "net/http"

// HTTPMiddleware interface serves as the building block for composing
// http middleware.
type HTTPMiddleware interface {
	// Register middleware
	Register(http.Handler) http.Handler
}
