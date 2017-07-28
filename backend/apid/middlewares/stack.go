package middlewares

import "net/http"

// Stack of middleware; stack itself is middleware and can but composed along
// with other middleware / stacks.
//
// eg. Apply(router, NewStack(Auth{}, BasicAuth{}, NewStack(Logger{})), ...)
type Stack struct {
	middleware []HTTPMiddleware
}

// NewStack returns new middleware stack
func NewStack(ms []HTTPMiddleware) *Stack {
	return &Stack{middleware: ms}
}

// Then ...
func (m Stack) Then(next http.Handler) http.Handler {
	return Apply(next, m.middleware...)
}
