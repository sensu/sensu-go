package middlewares

import "net/http"

type HTTPMiddleware interface {
	Register(http.Handler) http.Handler
}
